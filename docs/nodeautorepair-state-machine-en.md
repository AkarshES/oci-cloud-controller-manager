# Node AutoRepair State Machine Design

## Purpose
Node AutoRepair uses a state machine to repair a node after it has stayed continuously unhealthy for at least 10 minutes. The repair sequence is Cordon -> Drain -> Reboot -> Uncordon. The goals are idempotency, recoverability, and observability.

## Storage
- Use Node annotations as the lightweight storage for state (no CRD required). Suggested annotation keys:
  - `oci.oraclecloud.com/nodeautorepair-state` — current state (Detected/Cordoning/Draining/Rebooting/Uncordoning/Succeeded/Failed)
  - `oci.oraclecloud.com/nodeautorepair-repair-id` — unique repair task id to avoid concurrent repairs
- `oci.oraclecloud.com/nodeautorepair-last-transition` — ISO8601 timestamp of last transition
- `oci.oraclecloud.com/nodeautorepair-attempts` — per-state discrete action attempts
- `oci.oraclecloud.com/nodeautorepair-cycle-attempts` — failed repair cycles without an intervening healthy recovery
- `oci.oraclecloud.com/nodeautorepair-unhealthy-since` — start timestamp for the current continuous unhealthy dwell window
- `oci.oraclecloud.com/nodeautorepair-last-repair-end` — end timestamp of the last completed repair cycle
- `oci.oraclecloud.com/nodeautorepair-last-result` — last completed repair result (`succeeded` or `failed`)
- `oci.oraclecloud.com/node-auto-repair-human-intervention=true` label — hard stop after 3 consecutive failed repair cycles; must be cleared manually

## State Definitions
- Detected
- Cordoning
- Draining
- Rebooting
- Uncordoning
- Succeeded
- Failed

## State Transitions (Core Reconcile Flow)
1. Reconcile fetches the Node and evaluates unhealthy conditions.
2. If the node is healthy, cleanup transient repair annotations and any stale `unhealthy-since` marker.
3. If the node is unhealthy and no repair is in progress, require a continuous unhealthy dwell window of 10 minutes before starting a repair cycle.
4. If the node carries `oci.oraclecloud.com/node-auto-repair-human-intervention=true`, do not start a new repair cycle.
5. Once the dwell window is satisfied, create/continue a repair and switch on state:
2. switch on state:
   - Detected -> write state `Cordoning` (and record `repair-id`)
   - Cordoning -> perform cordon(node); on success set state `Draining`; on failure retry or set `Failed`
   - Draining -> perform drain(node); on success set state `Rebooting`; on failure retry or set `Failed`
   - Rebooting -> call cloud reboot API (or annotate node for agent); on success set state `Uncordoning`; otherwise retry/fail
   - Uncordoning -> in the normal success path, wait for node health to clear and stabilize before uncordon
   - Failed cleanup path -> if a repair step fails after NAR already cordoned the node, keep retrying uncordon until the node is schedulable again, then finalize the cycle as `failed`
   - Succeeded/Failed -> noop after terminal annotations have been finalized

Every state write should update `last-transition` and `attempts`.

## Idempotency and Retry
- All operations must be idempotent: repeated cordon/uncordon/drain calls should not cause inconsistent state.
- Default retry policy for discrete API actions is max 3 attempts with exponential backoff (base=10s). This applies to Cordoning, reboot submission, and the normal Uncordoning path.
- Repair cycles are retried separately from per-state action attempts:
  - If a repair cycle fails, clear transient repair state and wait for a fresh 10-minute continuous unhealthy window before retrying.
  - Retry up to 3 failed repair cycles.
  - After the 3rd failed cycle, label the node with `oci.oraclecloud.com/node-auto-repair-human-intervention=true` and stop automatic repair until an operator clears the label.
- Per-state timeouts (current defaults in code): Cordoning 60s, Draining 15m, Rebooting 20m, Uncordoning 5m. These are configurable via:
  - `NODE_AUTOREPAIR_TIMEOUT_CORDONING`
  - `NODE_AUTOREPAIR_TIMEOUT_DRAINING`
  - `NODE_AUTOREPAIR_TIMEOUT_REBOOTING`
  - `NODE_AUTOREPAIR_TIMEOUT_UNCORDONING`
- Additional timing defaults from current code:
  - unhealthy dwell threshold before starting repair: `10m` via `NODE_AUTOREPAIR_UNHEALTHY_THRESHOLD`
  - reboot polling interval: `20s` via `NODE_AUTOREPAIR_REBOOT_POLL_INTERVAL`
  - healthy stabilization before successful uncordon: `75s` via `NODE_AUTOREPAIR_HEALTHY_STABILIZATION`
  - drain force-after threshold: `10m` via `NODE_AUTOREPAIR_DRAIN_FORCE_AFTER`

## Safety Constraints
- Before Draining, respect PodDisruptionBudgets (PDB). If PDB blocks eviction, wait and retry. Respect PDB for a maximum of 10 mins and force repair after the wait
- Skip DaemonSet pods and mirror pods when draining; handle local-volume usage carefully (see kubectl drain behavior).
- Only the leader instance performs active repairs (use existing leader election).
- Concurrency limits:
  - Per-node: use `repair-id` in annotations to prevent concurrent repairs on the same node.
  - Global: only repair one node at any time, don't repair multiple machines at the same time, consider apply a controller wide lock to guarantee even if multiple node are with unhealthy conditions, only one node will be picked for repair at a time
- There could be multiple node auto repair controller in a cluster, please add leader election during controller initialization and make sure
only one controller is activelly doing node auto repair
- For global serialization add a cluster-scoped lock (e.g., `coordination.k8s.io/v1` Lease in `kube-system`) so only a single node repair runs at any time even if multiple controllers are reconciling. Plase make sure lease ownership is doesn't drift 
- Allow customer to exempt a node for auto repair by adding "oci.oraclecloud.com/node-auto-repair-disabled" label with value true. When present, the controller logs detected issues, skips cordon/drain/reboot, and simply waits for either the label or the node's conditions to change before reconciling again (no periodic requeue spam).

## Node Opt-Out Label
- Nodes labeled with `oci.oraclecloud.com/node-auto-repair-disabled=true` are treated as explicitly opted out of automated remediation.
- The controller will not start a new repair cycle while the label remains.
- If a repair is already in progress, the state machine continues and completes cleanup even if this label is added mid-repair.
- No periodic requeue is scheduled for opted-out nodes; reconciliation resumes automatically when either the node's conditions change or the label is removed/updated.
- Updates to this label are observed directly by the controller so operators can toggle opt-out without forcing artificial condition changes.

## Implementation Recommendations
- Node should be repaired serially, Each time the repair controller should only repair a node
- Start repair only after the node has been continuously unhealthy for 10 minutes.
- Retry failed repair cycles only after another continuous 10-minute unhealthy dwell window.
- After 3 consecutive failed repair cycles, add the human-intervention label and stop until an operator clears it.
- Cordon/Uncordon: update `Node.Spec.Unschedulable` via `client-go` (idempotent).
- Drain: prefer reusing `k8s.io/kubectl/pkg/drain`'s `drain.Helper` to correctly handle PDBs, DaemonSets, and local PVs. If not possible, implement eviction via the Eviction subresource and wait for pods to terminate while respecting PDB. Respect PDB for a maximum of 10 mins and force repair after the wait
- Reboot: reuse existing OCI client in `pkg/oci` to call instance reboot APIs; After reboot, we will check if instance is up and running using polling, if instance is not up and running, we wait for `instanceRunningPollInterval` (default 20s, configurable via `NODE_AUTOREPAIR_REBOOT_POLL_INTERVAL`) until instance is running again, if after 20 mins instance is not running, we move to failed step
- Uncordon success path: after reboot, if unhealthy conditions persist, keep waiting and only uncordon once the node is healthy and has stayed healthy for the stabilization window.
- Uncordon failed-cycle cleanup path: if a repair cycle fails after NAR already cordoned the node, uncordon even if unhealthy conditions still persist, and only then record the cycle as failed.
- Annotation updates should use optimistic concurrency and retry on resourceVersion conflicts.
- If the node is healthy again, it should be uncordoned
- If the repair failed, remove transient in-progress annotations so the next repair cycle starts cleanly; keep only summary information and the failed-cycle counter when more retries remain.

## Observability and Alerts
- Emit Kubernetes Events for state transitions and failures.
- Export Prometheus metrics: `nodeautorepair_repair_total`, `nodeautorepair_repair_failures_total`, `nodeautorepair_repair_duration_seconds` (by state)

## Testing Strategy
- Unit tests: state handler success, failure, and idempotency paths.
- Integration tests: mock API server and cloud reboot API to validate full path from `Detected` to `Succeeded`.

## Logging and Annotations
- Keep `repair-id`, `last-transition`, and `attempts` in node annotations.
- Logs should include nodeName, repair-id, state, error, attempts, and timestamps.

## Failure Handling and Human Intervention
- A repair cycle is only considered fully failed after cleanup is complete and the node is no longer left intentionally cordoned by NAR.
- If a repair step fails after NAR has cordoned the node, keep retrying uncordon until it succeeds.
- After the first and second failed repair cycles, wait for another fresh 10-minute unhealthy dwell window before retrying automatically.
- After the third failed repair cycle, emit failure events, label the node for human intervention, and stop automatic retries.

## Backwards Compatibility and Migration
- If annotations are absent, the controller should default to the legacy detection path and create an initial `Detected` entry.
- Document annotation keys and configuration knobs.

## Suggested Implementation Steps
1. Add state machine runner and state enums in `controllers/` (skeleton).
2. Implement `Cordoning` and `Draining` with unit tests.
3. Implement `Rebooting` using `pkg/oci`, then `Uncordoning`.
4. Add metrics, events, tests and documentation.
5. After each repair cycle succeeds or fails, cleanup transient annotations, preserve repair summary fields, and preserve failed-cycle count only while automatic retries are still allowed

---

This file documents the state machine design and implementation guidance for refactoring `controllers/nodeautorepair_controller.go`.
