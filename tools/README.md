tree recovery  --service committer --batch 1000 --height 4 --config ./tools/recovery/etc/config.yaml

tree recovery  --service witness --batch 1000 --height 26 --config ./tools/recovery/etc/config.yaml

treedb query  --service witness --height 54 --config ./tools/query/etc/config.yaml

treedb query  --service committer --height 54 --config ./tools/query/etc/config.yaml

revertblock --config ./tools/revertblock/etc/config.yaml --height 2

estimategas --config ./tools/estimategas/etc/config.yaml --fromHeight 196 --toHeight 238 --maxBlockCount 43  --sendToL1 1


rollback --config ./tools/rollback/etc/config.yaml --height 5

rollbackwitnesssmt --config ./tools/rollbackwitnesssmt/etc/config.yaml --height 2

redis-cli -h 127.0.0.1 -p 6666 flushdb