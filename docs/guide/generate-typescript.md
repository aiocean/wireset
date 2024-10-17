# Generate TypeScript types from Go

You can generate TypeScript types from Go structs using the `tygo` command.

**1. Install `tygo`:**

```bash
go install github.com/tygo/tygo/cmd/tygo@latest
```

**2. Create a `tygo.yaml` configuration file:**

```yaml:tygo.yaml
packages:
  - path: ./pkg/models
    types:
      - User
```

This configuration tells `tygo` to generate TypeScript types for the `User` struct in the `./pkg/models` package.

**3. Run `tygo` to generate the types:**

```bash
tygo generate
```

This will create a `types.ts` file in the same directory as your Go files, containing the generated TypeScript types.

**Example:**

**Go struct (`pkg/models/user.go`):**

```go:pkg/models/user.go
package models

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}
```

**Generated TypeScript types (`types.ts`):**

```typescript:types.ts
export interface User {
  id: number;
  firstName: string;
  lastName: string;
  email: string;
}