

```shell
norn pick -v <vendor> -r <repo> -s <sha> --token <token> --for <source ref>
# pick
norn pick \
    -v <vendor> \
    -r <repo> \
    -s <sha> \
    --token <token> \
    --merge-request-id <pull request id> \
    --for <source ref>

# summary for mr
norn pick \
    -v <vendor> \
    -r <repo> \
    -s <sha> \
    --token <token> \
    --merge-request-id 54 \
    --is-summary
```

```yaml
branches:
 - b1
 - b2
 - m
```