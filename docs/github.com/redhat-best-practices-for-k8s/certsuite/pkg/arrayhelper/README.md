## Package arrayhelper (github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper)



### Functions

- **ArgListToMap** — func([]string)(map[string]string)
- **FilterArray** — func([]string, func(string) bool)([]string)
- **Unique** — func([]string)([]string)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  ArgListToMap --> make
  ArgListToMap --> ReplaceAll
  ArgListToMap --> Split
  ArgListToMap --> len
  FilterArray --> make
  FilterArray --> f
  FilterArray --> append
  Unique --> make
  Unique --> make
  Unique --> len
  Unique --> append
```

### Symbol docs

- [function ArgListToMap](symbols/function_ArgListToMap.md)
- [function FilterArray](symbols/function_FilterArray.md)
- [function Unique](symbols/function_Unique.md)
