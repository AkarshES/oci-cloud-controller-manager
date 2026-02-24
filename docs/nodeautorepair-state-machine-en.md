# Node AutoRepair State Machine Design

## Purpose
Refactor the existing Node AutoRepair controller into a state machine that performs the following sequence for a detected unhealthy node: Cordon -> Drain -> Reboot -> Uncordon. The goals are idempotency, recoverability, and observability.

## Storage
- Use Node annotations as the lightweight storage for state (no CRD required). Suggested annotation keys:
  - `oci.oraclecloud.com/nodeautorepair-state` — current state (Detected/Cordoning/Draining/Rebooting/Uncordoning/Succeeded/Failed)
  - `oci.oraclecloud.com/nodeautorepair-repair-id` — unique repair task id to avoid concurrent repairs
  - `oci.oraclecloud.com/nodeautorepair-last-transition` — ISO8601 timestamp of last transition
  - `oci.oraclecloud.com/nodeautorepair-attempts` — number of attempts

## State Definitions
- Detected
- Cordoning
- Draining
- Rebooting
- Uncordoning
- Succeeded
- Failed

## State Transitions (Core Reconcile Flow)
1. Reconcile fetches the Node and reads state from annotations (default to `Detected` if absent)
2. switch on state:
   - Detected -> write state `Cordoning` (and record `repair-id`)
   - Cordoning -> perform cordon(node); on success set state `Draining`; on failure retry or set `Failed`
   - Draining -> perform drain(node); on success set state `Rebooting`; on failure retry or set `Failed`
   - Rebooting -> call cloud reboot API (or annotate node for agent); on success set state `Uncordoning`; otherwise retry/fail
   - Uncordoning -> perform uncordon(node); on success set state `Succeeded`; otherwise retry/fail
   - Succeeded/Failed -> noop (preserve history and emit events for manual remediation)

Every state write should update `last-transition` and `attempts`.

## Idempotency and Retry
- All operations must be idempotent: repeated cordon/uncordon/drain calls should not cause inconsistent state.
- Default retry policy: max 3 attempts with exponential backoff (base=10s).
- Per-state timeouts: Cordoning 30s, Draining 10m, Rebooting 5m, Uncordoning 30s (configurable).

## Safety Constraints
- Before Draining, respect PodDisruptionBudgets (PDB). If PDB blocks eviction, delay retry or mark `Failed` (configurable behavior).
- Skip DaemonSet pods and mirror pods when draining; handle local-volume usage carefully (see kubectl drain behavior).
- Only the leader instance performs active repairs (use existing leader election).
- Concurrency limits:
  - Per-node: use `repair-id` in annotations to prevent concurrent repairs on the same node.
  - Global: provide a configurable limit for concurrent repairs.

## Implementation Recommendations
- Cordon/Uncordon: update `Node.Spec.Unschedulable` via `client-go` (idempotent).
- Drain: prefer reusing `k8s.io/kubectl/pkg/drain`'s `drain.Helper` to correctly handle PDBs, DaemonSets, and local PVs. If not possible, implement eviction via the Eviction subresource and wait for pods to terminate while respecting PDB.
- Reboot: reuse existing OCI client in `pkg/oci` to call instance reboot APIs; as a fallback annotate the node and let a node agent perform the reboot.
- Annotation updates should use optimistic concurrency and retry on resourceVersion conflicts.

## Observability and Alerts
- Emit Kubernetes Events for state transitions and failures.
- Export Prometheus metrics: `nodeautorepair_repair_total`, `nodeautorepair_repair_failures_total`, `nodeautorepair_repair_duration_seconds` (by state)

## Testing Strategy
- Unit tests: state handler success, failure, and idempotency paths.
- Integration tests: mock API server and cloud reboot API to validate full path from `Detected` to `Succeeded`.
- Optional e2e: inject unhealthy node into a cluster and observe an end-to-end repair.

## Logging and Annotations
- Keep `repair-id`, `last-transition`, and `attempts` in node annotations.
- Logs should include nodeName, repair-id, state, error, attempts, and timestamps.

## Failure Handling and Human Intervention
- When a repair reaches `Failed`, preserve annotations and emit an Event for operators.
- Provide configurable automated retry or surface failures for manual remediation.

## Backwards Compatibility and Migration
- If annotations are absent, the controller should default to the legacy detection path and create an initial `Detected` entry.
- Document annotation keys and configuration knobs.

## Suggested Implementation Steps
1. Add state machine runner and state enums in `controllers/` (skeleton).
2. Implement `Cordoning` and `Draining` with unit tests.
3. Implement `Rebooting` using `pkg/oci`, then `Uncordoning`.
4. Add metrics, events, tests and documentation.

---

This file documents the state machine design and implementation guidance for refactoring `controllers/nodeautorepair_controller.go`.