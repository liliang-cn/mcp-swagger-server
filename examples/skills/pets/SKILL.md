---
name: Pets
description: API operations for pets (6 endpoints)
---

# Pets API Skill

This skill provides tools for interacting with the pets API.

## When to use this skill

Use this skill when you need to:
- **listPets**: List all pets
- **createPet**: Create a pet
- **getPetById**: Get a pet by ID
- **updatePet**: Update a pet
- **deletePet**: Delete a pet
- **searchPets**: Search for pets

## Available Tools

### listpets
**GET /pets**

Returns a list of all pets in the store with optional filtering

### createpet
**POST /pets**

Create a new pet in the store

### getpetbyid
**GET /pets/{petId}**

Returns a single pet

### updatepet
**PUT /pets/{petId}**

Update an existing pet by ID

### deletepet
**DELETE /pets/{petId}**

Delete a pet by ID

### searchpets
**POST /pets/search**

Search for pets using various criteria

## Configuration

- **Base URL**: `http://localhost:8080`

## Additional Documentation

See [reference.md](reference.md) for detailed API documentation including:
- Request/response schemas
- Parameter descriptions
- Error handling
- Usage examples
