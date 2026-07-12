---
title: Read-path performance follow-up plan
type: plan
status: active
created: "2026-07-12T20:06:25.019823Z"
updated: "2026-07-12T20:10:58.240862Z"
tasks:
    - T-BATCH-SCRATCHPAD-DOCUMENT-REFERENCE
    - T-BENCHMARK-READ-PATHS-BEFORE
    - T-ELIMINATE-REDUNDANT-STORE-LOADS
    - T-INDEX-TASK-COLUMNS-PER
---
# Read-path performance follow-ups

1. Add snapshot-level helpers so document, task, and index rendering does not reload the file store for each related value. Keep the public data immutable by returning value copies.
2. Add benchmarks for board, document, and task view assembly at representative store sizes. Use them to decide whether a cache offers a material benefit.
3. Validate all scratchpad document references from one snapshot, preserving the current request errors and rollback behavior.
4. Build a task-to-column index as part of load repair and use it for task conversion without changing board ordering or repair semantics.
5. Run the full race-enabled check and record benchmark results. Do not add an in-memory cache unless it has a correct external-change invalidation contract.

## Results

The read-view benchmarks measure a fresh file-store load with linked synthetic entities. On an Apple M1 Pro, a one-iteration run at 1,000 tasks and 1,000 documents measured board at 55.5 ms, document at 85.0 ms, and task at 52.6 ms. The implementation deliberately does not add a cache: these results meet the sub-100 ms target while preserving immediate visibility of external CLI and Git edits.
