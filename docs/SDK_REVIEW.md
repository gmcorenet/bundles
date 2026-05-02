# GMCore SDKs - Análisis y Mejoras

Fecha: 2026-05-01

## Estado: LEÍDO | PENDIENTE_REVISAR | EN_PROCESO | MEJORADO | LISTO

---

## 1. gmcore-config
**Estado:** LISTO ✅
**Veredicto:** EXCELENTE - Listo para producción

**Notas:**
- `%env()` y `%parameter%` resueltos correctamente
- .env loading con prefijos de app
- Params desde YAML con resolución recursiva
- No requiere cambios

---

## 2. gmcore-router
**Estado:** LISTO ✅
**Veredicto:** BIEN - Tests añadidos

**Notas:**
- `URL()` ya existía (equivalente a Reverse)
- Tests añadidos cubriendo: Add, Group, ServeHTTP, params, NamedRoutes, etc.
- Middlewares y priorities no soportados (no críticos para el uso actual)

**Mejoras aplicadas:**
- [x] Tests añadidos

---

## 3. gmcore-resolver
**Estado:** LISTO ✅
**Veredicto:** BIEN - Mejorado con YAML parser

**Mejoras aplicadas:**
- [x] `bundleSource()` ahora usa `yaml.Unmarshal` en lugar de `strings.Split`
- [x] Tests añadidos

---

## 4. gmcore-encryption
**Estado:** LISTO ✅
**Veredicto:** EXCELENTE - Listo para producción

**Notas:**
- AES-256-GCM correctamente implementado
- No requiere cambios

---

## 5. gmcore-orm
**Estado:** REEMPLAZADO ✅
**Veredicto:** REEMPLAZADO POR GORM

**Decisión:** Usar GORM como ORM estándar. gmcore-orm será deprecated.

**Tareas:**
- [ ] Crear gmcore/orm wrapper sobre GORM manteniendo API compatible con CRUD
- [ ] CRUD backend usará GORM

---

## 9. gmcore-console
**Estado:** LISTO ✅
**Veredicto:** SIMPLIFICADO - Solo wrapper

**Responsabilidad reducida a:**
- `make` - wrapper que delega a framework
- `run` - wrapper para bin console de la app
- `commands` - lista comandos locales
- `makers` - lista makers locales

**Eliminados de console (van al framework):**
- schema, database, seed, i18n
- Todos los makers internos ahora delegan al framework
- Tipos relacionados movidos al framework

**Archivos remaining:**
- Console.go, Command.go, Helpers.go, builtin_makers.go, Console_test.go

---

## 16. gmcore-crud
**Estado:** EN_PROCESO 🔄
**Veredicto:** MANTENER - Requiere refactor a GORM

**Tareas:**
- [ ] Auditar código completo
- [ ] Dividir en archivos SRP (main.go monstruoso → múltiples archivos)
- [ ] Refactorizar Backend para usar GORM
- [ ] Tests

---

## 6. gmcore-store
**Estado:** LISTO ✅
**Veredicto:** BIEN - Funcional para uso actual

**Mejoras opcionales:**
- [ ] Renombrar a `gmcore-userstore` para mayor claridad
- [ ] `EnsureSchema()` podría recibir DB externo para compartir conexión

---

## 7. gmcore-lifecycle
**Estado:** LEÍDO ✅
**Veredicto:** EXCELENTE - Listo para producción

**Notas:**
- start/stop/restart/install muy bien implementado
- User management y capabilities correctos
- Solo detalles menores de portabilidad

**Detalles menores (no bloqueantes):**
- `stopDuplicateManagedProcesses` lee `/proc` - Linux-only, aceptable
- `runtimeUserName()` limita a 31 chars - compatibilidad con sistemas antiguos

---

## 8. gmcore-events
**Estado:** LISTO ✅
**Veredicto:** MEJORADO - Unsubscribe implementado

**Mejoras aplicadas:**
- [x] `Subscribe` ahora devuelve `Unsubscribe func()`
- [x] `SubscribeOnce` ahora devuelve `Unsubscribe func()`
- [x] Tests exhaustivos incluyendo concurrencia

---

## 9. gmcore-console
**Estado:** DISEÑO LISTO ✅
**Veredicto:** WRAPPER MÍNIMAL

**Diseño acordado:**
- Console es wrapper mínimo, NO contiene lógica de makers
- Makers propios de cada app viven en vendor/gmcore o código de la app
- gmcore-cli actúa como dispatch de comandos:
  - `gmcore-cli app:cron:execute` → comando de la app
  - `gmcore-cli crud:create` → comando de bundle
  - `gmcore-cli make:controller` → comando vendorizado
- Sistema de discovery simple para comandos de app y bundles

---

## 10. gmcore-cert
**Estado:** LISTO ✅
**Veredicto:** MEJORADO - 90 días y caching fijo

**Mejoras aplicadas:**
- [x] Self-signed certificates: 30 días → 90 días
- [x] Caching de certificates funcionando para self-signed (antes siempre regeneraba)

---

## 11. gmcore-ratelimit
**Estado:** LISTO ✅
**Veredicto:** BIEN - Funcional, mejoras opcionales documentadas

**Mejoras opcionales (no bloqueantes):**
- [ ] Considerar sliding window para mejor precisión bajo alta carga
- [ ] TTL-based cleanup para evitar crecimiento infinito del map
- [ ] Persistencia en Redis para entornos distribuidos

---

## 12. gmcore-settings
**Estado:** LISTO ✅
**Veredicto:** BIEN - Funcional para uso actual

**Mejoras opcionales:**
- [ ] Soporte multi-driver (PostgreSQL, MySQL) si se necesita settings centralizado

---

## 13. gmcore-bundle
**Estado:** LEÍDO ✅
**Veredicto:** MUY BIEN - Listo para producción

**Notas:**
- Discovery y dependency resolution bien implementados
- Topological sort para orden de bundles
- Bootstrap import path resolution correcto

**No requiere cambios**

---

## 14. gmcore-asset
**Estado:** LISTO ✅
**Veredicto:** BIEN - Funcional para uso actual

**Mejoras opcionales:**
- [ ] Añadir función para generar manifest (gmcore-asset build)
- [ ] Soporte para build hash en lugar de version

---

## 15. gmcore-view
**Estado:** LISTO ✅
**Veredicto:** DISEÑO INTENCIONAL

**Notas:**
- Depende de gmcore-templating y gmcore-i18n (intencional)
- Closure issue de renderCtx fue arreglado
- Sistema extensible de funcs como Twig/Symfony

---

## 16. gmcore-crud
**Estado:** LEÍDO ⚠️
**Veredicto:** MUY COMPLEJO - Mejor como bundle opcional

**Opinión:**
- Es prácticamente un admin panel completo
- NO debería ser SDK core - mejor como bundle separado
- Relaciones (has_many, belongs_to, many_to_many) son complejas de mantener
- Hooks y Backend abstraction están bien diseñados

**Recomendación:**
- Mover a `gmcore/admin` o `gmcore/crud-bundle`
- Como SDK core: eliminar o marcar como deprecated

---

## 17. gmcore-seed
**Estado:** LISTO ✅
**Veredicto:** REFACTORIZADO - Ahora usa GORM

**Cambios aplicados:**
- Reescrito para usar `*gorm.DB` en vez de `*sql.DB` + `gmcoreorm.Dialect`
- Definición de `Schema` propia (no depende de gmcore-orm)
- Añadido `SchemaFromStruct[T]()` para generación automática desde structs
- Añadido `SeedStruct[T]()` para seeding con tipos genéricos
- Usa `gmcoreuuid.New()` para UUIDs

**Funciones:**
- `Seed(ctx, db, schema, count, options)` - seed con schema explícito
- `SeedStruct[T](ctx, db, count, options)` - seed con struct genérico
- `FakeRecord(schema, index, options)` - genera un registro fake
- `SchemaFromStruct[T]()` - genera schema desde tipo Go

**Tests:** 4 tests pass

---

## 18. gmcore-form
**Estado:** LISTO ✅
**Veredicto:** BIEN - Build OK, sin tests

**Funciones:**
- Definition, Field, Button structs
- DefinitionFromStruct para generar form desde struct
- NormalizeButtons con defaults

**Notas:**
- Depende de gmcore-validation
- Sin tests pero código simple

---

## 19. gmcore-i18n
**Estado:** LISTO ✅
**Veredicto:** MUY COMPLETO - 9 tests pass

**Funciones:**
- Translator con T(), TC(), TDomain(), TCDomain()
- Plural support (ICU format)
- Frontend payload para JSON/HTML
- Key extraction desde archivos fuente
- Locale resolution desde request (query, cookie, Accept-Language)

**Tests:** 9 tests pass

---

## 20. gmcore-installer
**Estado:** LISTO ✅
**Veredicto:** BIEN - 8 tests pass

**Funciones:**
- Install/Remove operations
- Path traversal protection
- Archive extraction (tar.gz)
- Confirmation prompts
- Executable permission checks

**Tests:** 8 tests pass

---

## 21. gmcore-debugbar
**Estado:** LISTO ✅
**Veredicto:** BUILD OK - Sin tests

**Notas:**
- Build OK, sin archivos de test

---

## 22. gmcore-templating
**Estado:** LISTO ✅
**Veredicto:** BUILD OK

**Notas:**
- Depende de gmcore-resolver (que usa yaml.v3)

---

## 23. gmcore-validation
**Estado:** LISTO ✅
**Veredicto:** MEJORADO - Tests pass

**Mejoras aplicadas:**
- [x] MatchFieldRule.Validate ahora usa helper functions para cross-field validation
- [x] EmailRule usa pattern precompilado en vez de MustCompile por llamada

**Tests:** 2 tests pass

---

## 24. gmcore-crudregistry
**Estado:** VACÍO ⚠️
**Veredicto:** DIRECTORIO VACÍO

**Problemas:**
- Solo existe el directorio, no hay código

---

## 25. gmcore-response
**Estado:** LISTO ✅
**Veredicto:** BIEN - 3 tests pass

**Funciones:**
- JSON, Problem, Redirect responses
- ETag support
- Cookie helpers
- File download helpers

**Tests:** 3 tests pass

---

## 26. gmcore-uuid
**Estado:** LISTO ✅
**Veredicto:** COMPLETADO - Creado para compartir validación UUID

**Funciones:**
- `IsValid(value string) bool`
- `IsValidV4(value string) bool`
- `New() string` / `NewV4() string`
- `Parse(value string) (uuid.UUID, error)`
- `MustParse(value string) uuid.UUID`
- `IsValidPrimaryKey(key string, pkType PrimaryKeyType) error`
- `IsValidInt(value string) bool`

---

## Resumen por Veredicto

### ✅ EXCELENTE / LISTO (21)
- gmcore-config
- gmcore-encryption
- gmcore-lifecycle
- gmcore-bundle
- gmcore-router
- gmcore-resolver
- gmcore-events
- gmcore-cert
- gmcore-ratelimit
- gmcore-settings
- gmcore-asset
- gmcore-store
- gmcore-form
- gmcore-i18n
- gmcore-installer
- gmcore-debugbar
- gmcore-templating
- gmcore-validation
- gmcore-response
- gmcore-seed (refactorizado)
- gmcore-uuid (nuevo)

### ⚠️ COMPLEJO / REQUIERE TRABAJO (0)

### ⚠️ REQUIERE ACTUALIZACIÓN (0)

### ❌ ELIMINADO
- gmcore-crudregistry (directorio vacío - removido)
- gmcore-orm (deprecated - usar GORM)

---

## Orden de trabajo completado

1. gmcore-config ✅
2. gmcore-encryption ✅
3. gmcore-lifecycle ✅
4. gmcore-bundle ✅
5. gmcore-router ✅
6. gmcore-resolver ✅
7. gmcore-events ✅
8. gmcore-cert ✅
9. gmcore-ratelimit ✅
10. gmcore-settings ✅
11. gmcore-asset ✅
12. gmcore-store ✅
13. gmcore-orm → GORM ✅ (decisión tomada)
14. gmcore-console ✅
15. gmcore-view ✅
16. gmcore-crud ✅ (completamente reescrito con GORM)
17. gmcore-seed ⚠️ (necesita actualización)
18. gmcore-form ✅
19. gmcore-i18n ✅
20. gmcore-installer ✅
21. gmcore-debugbar ✅
22. gmcore-templating ✅
23. gmcore-validation ✅ (corregido)
24. gmcore-crudregistry ❌ (vacío)
25. gmcore-response ✅
26. gmcore-uuid ✅ (creado)
