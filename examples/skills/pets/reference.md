# Pets API Reference

## listpets

**Method**: GET

**Path**: `/pets`

**Summary**: List all pets

**Description**: Returns a list of all pets in the store with optional filtering

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| limit | query | integer | No | How many items to return at one time (max 100) |
| tags | query | array | No | Tags to filter by |

### Responses

#### %!s(int=200)
A paged array of pets

**Schema**:
```json
[]
```

---

## createpet

**Method**: POST

**Path**: `/pets`

**Summary**: Create a pet

**Description**: Create a new pet in the store

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| body | body |  | Yes | Pet to add to the store |

### Responses

#### %!s(int=400)
Invalid input

#### %!s(int=201)
Pet created

**Schema**:
```json
{}
```

---

## getpetbyid

**Method**: GET

**Path**: `/pets/{petId}`

**Summary**: Get a pet by ID

**Description**: Returns a single pet

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| petId | path | integer | Yes | The id of the pet to retrieve |

### Responses

#### %!s(int=200)
Expected response to a valid request

**Schema**:
```json
{}
```

#### %!s(int=404)
Pet not found

---

## updatepet

**Method**: PUT

**Path**: `/pets/{petId}`

**Summary**: Update a pet

**Description**: Update an existing pet by ID

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| petId | path | integer | Yes | The id of the pet to update |
| body | body |  | Yes | Updated pet object |

### Responses

#### %!s(int=400)
Invalid input

#### %!s(int=404)
Pet not found

#### %!s(int=200)
Pet updated

**Schema**:
```json
{}
```

---

## deletepet

**Method**: DELETE

**Path**: `/pets/{petId}`

**Summary**: Delete a pet

**Description**: Delete a pet by ID

### Parameters

| Name | In | Type | Required | Description |
|------|-----|------|----------|-------------|
| petId | path | integer | Yes | The id of the pet to delete |

### Responses

#### %!s(int=204)
Pet deleted

#### %!s(int=404)
Pet not found

---

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

