# Node Auto Repair Watcher

`hack/nodeautorepair_watch.py` is a one-shot watcher for validating `node auto repair` behavior against the current implementation in:

- `controllers/nodeautorepair_controller.go`
- `controllers/repair_statemachine.go`

It continuously samples a target node, collects raw evidence, and generates a Chinese validation report.

## What It Collects

Each sampling iteration records:

- `kubectl get node <node> -o json`
- `kubectl get events --all-namespaces --field-selector involvedObject.kind=Node,involvedObject.name=<node> -o json`
- `kubectl get pods -A --field-selector spec.nodeName=<node> -o json`

The watcher writes:

- `raw/node.jsonl`: full node snapshots
- `raw/events.jsonl`: full event snapshots
- `raw/pods.jsonl`: full pod snapshots for the node
- `timeline.json`: normalized state, event, and readiness timeline
- `summary.json`: machine-readable consistency result
- `report.md`: human-readable validation report

## Default Validation Rules

The watcher validates behavior against the current code defaults:

- unhealthy dwell: `10m`
- cordoning timeout: `60s`
- draining timeout: `30m`
- rebooting timeout: `20m`
- uncordoning timeout: `5m`
- post-reboot observation window: `90s`
- force drain: enabled by default

It classifies the observed repair as one of:

- `consistent_success`
- `consistent_failure`
- `inconsistent_or_stuck`

The watcher checks:

- repair state transitions
- repair-id stability
- `last-transition`, `last-result`, and `last-repair-end`
- repair taint lifecycle
- `spec.unschedulable`
- `Ready` condition and `node.kubernetes.io/unreachable` taints

## Usage

Run with the defaults from this repo's current investigation target:

```bash
python3 hack/nodeautorepair_watch.py
```

Run against a specific kubeconfig, node, and output directory:

```bash
python3 hack/nodeautorepair_watch.py \
  --kubeconfig /Users/penzhou/.kube/config-ams \
  --node 10.140.67.114 \
  --sample-seconds 10 \
  --timeout-seconds 3600 \
  --out-dir artifacts/nodeautorepair/manual-run
```

Pin the expected repair ID when validating a known in-flight repair:

```bash
python3 hack/nodeautorepair_watch.py \
  --expected-repair-id 36bb303c-93f1-4656-98cc-f39021ef9a34
```

## Test

Run the watcher unit tests with:

```bash
python3 -m unittest hack/nodeautorepair_watch_test.py
```

## Notes

- The watcher is designed for one-shot continuous observation, not long-term periodic automation.
- It uses `Node` annotations and Kubernetes `Events` as the primary sources of truth.
- It does not inspect controller pod environment variables; report timing explanations are based on the current code defaults.
