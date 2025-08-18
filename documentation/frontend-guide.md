### Basin Admin Frontend Guide

This guide helps you bootstrap a Next.js admin UI alongside the Go backend in a monorepo.

## Repository layout

```
basin/          # Go API (this repo)
basin-admin/    # Next.js admin UI (new)
```

## Tech stack
- Next.js (App Router) + TypeScript
- Tailwind CSS + shadcn/ui
- TanStack Query (React Query)
- OpenAPI-generated client from `/swagger/doc.json`

## Create the app
```bash
# from repo root
npm create next-app basin-admin --ts --eslint --app --src-dir --tailwind
cd basin-admin
npm add @tanstack/react-query zod axios @zodios/core
npm add -D openapi-typescript
```

## Generate API types
```bash
# from repo root
npx openapi-typescript http://localhost:8080/swagger/doc.json -o basin-admin/src/lib/api-types.ts

# This will generate TypeScript types from the Swagger spec
# The types will automatically update when the API changes
# You can also view the API docs at http://localhost:8080/swagger/index.html
```

## Auto-sync with Swagger (Recommended)
Create a script to automatically regenerate types when the API changes:

```bash
# basin-admin/scripts/sync-api.sh
#!/bin/bash
echo "🔄 Syncing API types from Swagger docs..."
npx openapi-typescript http://localhost:8080/swagger/doc.json -o src/lib/api-types.ts
echo "✅ API types updated!"
```

```json
// basin-admin/package.json
{
  "scripts": {
    "sync-api": "bash scripts/sync-api.sh",
    "dev": "npm run sync-api && next dev",
    "build": "npm run sync-api && next build"
  }
}
```

## Manual API Type Updates
Whenever you modify the backend API:

1. **Update Swagger annotations** in your Go code
2. **Regenerate Swagger docs**: `make start` (auto-runs `swag init`)
3. **Sync frontend types**: `npm run sync-api`
4. **Restart frontend dev server** to pick up new types

## Available Swagger Endpoints
- **Interactive UI**: `http://localhost:8080/swagger/index.html`
- **JSON Spec**: `http://localhost:8080/swagger/doc.json` 
- **YAML Spec**: `http://localhost:8080/swagger/doc.yaml`

## Basic API client
```ts
// basin-admin/src/lib/api.ts
import axios from 'axios'
import type { paths } from './api-types'

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  withCredentials: true,
})

// Type-safe API calls using generated types
export const authAPI = {
  login: (data: paths['/auth/login']['post']['requestBody']['content']['application/json']) =>
    api.post<paths['/auth/login']['post']['responses']['200']['content']['application/json']>('/auth/login', data),
  
  me: () =>
    api.get<paths['/auth/me']['get']['responses']['200']['content']['application/json']>('/auth/me'),
}

export const itemsAPI = {
  list: (table: string, params?: {
    limit?: number
    offset?: number
    sort?: string
    order?: 'ASC' | 'DESC'
  }) =>
    api.get<paths['/items/{table}']['get']['responses']['200']['content']['application/json']>(`/items/${table}`, { params }),
  
  get: (table: string, id: string) =>
    api.get<paths['/items/{table}/{id}']['get']['responses']['200']['content']['application/json']>(`/items/${table}/${id}`),
  
  create: (table: string, data: Record<string, any>) =>
    api.post<paths['/items/{table}']['post']['responses']['201']['content']['application/json']>(`/items/${table}`, data),
  
  update: (table: string, id: string, data: Record<string, any>) =>
    api.put<paths['/items/{table}/{id}']['put']['responses']['200']['content']['application/json']>(`/items/${table}/${id}`, data),
  
  delete: (table: string, id: string) =>
    api.delete<paths['/items/{table}/{id}']['delete']['responses']['200']['content']['application/json']>(`/items/${table}/${id}`),
}
```

## Auth flow
- Call `POST /auth/login` with email/password.
- Store JWT in httpOnly cookie via a Next.js Route Handler (`/api/session/login`).
- Forward cookie to API on server actions/fetches.

## Pages to start
- `/login`: form → POST to `/api/session/login` → redirect to `/collections`
- `/collections`: list from `GET /items/collections?limit=50&sort=created_at&order=desc`
- `/collections/[name]`: show fields and data grid
- `/users`, `/roles`, `/permissions`: basic CRUD shells

## UI patterns
- Use shadcn/ui for forms, modals, tables.
- Use React Query for caching and mutations.
- Use Zod to validate forms.

## Using Generated Types in Components
```tsx
// basin-admin/src/components/CollectionsList.tsx
import { useQuery } from '@tanstack/react-query'
import { itemsAPI } from '@/lib/api'
import type { paths } from '@/lib/api-types'

type Collection = paths['/items/collections']['get']['responses']['200']['content']['application/json']['data'][0]

export function CollectionsList() {
  const { data, isLoading } = useQuery({
    queryKey: ['collections'],
    queryFn: () => itemsAPI.list('collections'),
  })

  if (isLoading) return <div>Loading...</div>

  return (
    <div>
      {data?.data.map((collection: Collection) => (
        <div key={collection.id}>
          <h3>{collection.name}</h3>
          <p>{collection.description}</p>
        </div>
      ))}
    </div>
  )
}
```

## Type-Safe Forms
```tsx
// basin-admin/src/components/CreateCollectionForm.tsx
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { itemsAPI } from '@/lib/api'
import type { paths } from '@/lib/api-types'

type CreateCollectionData = paths['/items/collections']['post']['requestBody']['content']['application/json']

export function CreateCollectionForm() {
  const [formData, setFormData] = useState<CreateCollectionData>({
    name: '',
    description: '',
    icon: '📁',
  })

  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: (data: CreateCollectionData) => itemsAPI.create('collections', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate(formData)
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={formData.name}
        onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
        placeholder="Collection name"
      />
      {/* More form fields */}
    </form>
  )
}
```

## Env
Create `basi n-admin/.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## CORS
Backend includes permissive CORS for dev. In production, restrict to your admin domain.

## Development Workflow
1. **Backend Changes**: Modify Go code with Swagger annotations
2. **Regenerate Docs**: `make start` (auto-runs `swag init`)
3. **Frontend Sync**: `npm run sync-api` (or restart dev server)
4. **Type Safety**: Enjoy full TypeScript intellisense and compile-time checks

## Development Tips
- Use the Swagger UI at http://localhost:8080/swagger/index.html to:
  - Explore available endpoints
  - Test API calls directly
  - View request/response schemas
  - Check authentication requirements
- The API types will automatically update when you:
  - Add new endpoints
  - Modify existing endpoints
  - Run `make start` or `make docs` on the backend

## Troubleshooting

### Types not updating?
```bash
# Check if Swagger endpoint is working
curl http://localhost:8080/swagger/doc.json

# Regenerate types manually
npm run sync-api

# Check generated file
cat src/lib/api-types.ts | head -20
```

### Type errors after API changes?
1. Ensure backend is running: `make start`
2. Verify Swagger docs are updated: `/swagger/index.html`
3. Regenerate types: `npm run sync-api`
4. Restart TypeScript server in your editor

### CORS issues?
- Backend includes permissive CORS for development
- Check browser console for CORS errors
- Verify `NEXT_PUBLIC_API_URL` is correct

## Next steps
- Add role/permission editor pages
- Add invite flow (backend endpoints TBD) and tenant switcher
- Add optimistic updates for data tables
- Consider using a tool like `@zodios/core` for type-safe API calls
