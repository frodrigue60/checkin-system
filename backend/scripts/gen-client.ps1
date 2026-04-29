# Generate TypeScript client from swagger.json
# 1. Convert Swagger 2.0 to OpenAPI 3.0
npx swagger2openapi ./docs/swagger.json --outfile ./docs/openapi3.json --patch

# 2. Generate TS types from the converted spec
npx openapi-typescript ./docs/openapi3.json -o ../frontend/src/lib/api/client/schema.d.ts
