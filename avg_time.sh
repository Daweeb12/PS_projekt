#!/usr/bin/env bash

RUNS=10
TOTAL=0

for i in $(seq 1 $RUNS); do
    echo "Run $i..."
    START=$(date +%s.%N)

    ./bin/craqSim >/dev/null

    END=$(date +%s.%N)
    ELAPSED=$(echo "$END - $START" | bc)
    TOTAL=$(echo "$TOTAL + $ELAPSED" | bc)
done

AVG=$(echo "scale=4; $TOTAL / $RUNS" | bc)
echo "Average time over $RUNS runs: $AVG seconds"

