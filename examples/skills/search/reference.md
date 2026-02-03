# Search API Reference

## searchpets

**Method**: POST

**Path**: `/pets/search`

**Summary**: Search for pets

**Description**: Search for pets using various criteria

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| body | body |  | Yes | Search criteria |

### Responses

#### %!s(int=200)
Search results

**Schema**:
```json
[]
```

#### %!s(int=400)
Invalid search criteria

---

