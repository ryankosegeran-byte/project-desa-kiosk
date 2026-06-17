---
type: community
cohesion: 0.10
members: 34
---

# JWT Authentication

**Cohesion:** 0.10 - loosely connected
**Members:** 34 nodes

## Members
- [[.Create()_4]] - code - server/db/user_repo.go
- [[.FindByID()_5]] - code - server/db/user_repo.go
- [[.FindByUsername()]] - code - server/db/user_repo.go
- [[.GenerateAccessToken()]] - code - internal/auth/jwt.go
- [[.GenerateRefreshToken()]] - code - internal/auth/jwt.go
- [[.List()_3]] - code - server/db/user_repo.go
- [[.ListActivityLogs()]] - code - server/db/user_repo.go
- [[.LogActivity()]] - code - server/db/user_repo.go
- [[.Update()_1]] - code - server/db/user_repo.go
- [[.UpdateLastLogin()]] - code - server/db/user_repo.go
- [[.UpdatePassword()]] - code - server/db/user_repo.go
- [[.ValidateToken()]] - code - internal/auth/jwt.go
- [[.scanRow()_6]] - code - server/db/user_repo.go
- [[.scanRows()_6]] - code - server/db/user_repo.go
- [[Claims]] - code - internal/auth/jwt.go
- [[Context_17]] - code - server/db/user_repo.go
- [[DB_13]] - code - server/db/seeder.go
- [[DB_16]] - code - server/db/user_repo.go
- [[Duration]] - code - internal/auth/jwt.go
- [[JWTManager]] - code - internal/auth/jwt.go
- [[NewJWTManager()]] - code - internal/auth/jwt.go
- [[NewUserRepository()]] - code - server/db/user_repo.go
- [[RegisteredClaims]] - code - internal/auth/jwt.go
- [[Row_6]] - code - server/db/user_repo.go
- [[Rows_6]] - code - server/db/user_repo.go
- [[SeedServerData()]] - code - server/db/seeder.go
- [[User_1]] - code - server/db/user_repo.go
- [[UserActivityLog_1]] - code - server/db/user_repo.go
- [[UserRepository_1]] - code - server/db/user_repo.go
- [[jwt.go]] - code - internal/auth/jwt.go
- [[main()_3]] - code - server/cmd/server/main.go
- [[main.go_1]] - code - server/cmd/server/main.go
- [[seeder.go_1]] - code - server/db/seeder.go
- [[user_repo.go]] - code - server/db/user_repo.go

## Live Query (requires Dataview plugin)

```dataview
TABLE source_file, type FROM #community/JWT_Authentication
SORT file.name ASC
```

## Connections to other communities
- 2 edges to [[_COMMUNITY_Sync Handlers]]
- 2 edges to [[_COMMUNITY_Desa Repository]]
- 1 edge to [[_COMMUNITY_Community 30]]
- 1 edge to [[_COMMUNITY_Community 23]]
- 1 edge to [[_COMMUNITY_OCR Google Gemini]]
- 1 edge to [[_COMMUNITY_OCR Groq]]
- 1 edge to [[_COMMUNITY_OCR Mistral AI]]
- 1 edge to [[_COMMUNITY_OCR Provider Interface]]

## Top bridge nodes
- [[main()_3]] - degree 11, connects to 6 communities
- [[SeedServerData()]] - degree 5, connects to 1 community
- [[Claims]] - degree 4, connects to 1 community
- [[jwt.go]] - degree 4, connects to 1 community
- [[Duration]] - degree 4, connects to 1 community