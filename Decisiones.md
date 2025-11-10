## 1. Objetivo y alcance

El TP7 consolida todo el trabajo previo en un solo repo (`tp7ingsoft3`) con backend en Go, frontend en React y MongoDB como base de datos. El objetivo principal fue integrar métricas de calidad (coverage, análisis estático, E2E) dentro de un pipeline único de GitHub Actions que funcione como quality gate real antes de habilitar cualquier despliegue.

## 2. Stack elegido y por qué

| Capa               | Tecnología                   | Motivo de la elección                                                                                                                                            |
| ------------------ | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend            | **Go 1.22 + Gin**            | Go nos da concurrencia y binarios ligeros, Gin permite exponer REST rápido con buen rendimiento y middlewares simples para auth y manejo de errores.             |
| Base de datos      | **MongoDB**                  | El dominio se modela como documentos (usuarios y todos). Mongo encaja con Go gracias al driver oficial y soporta despliegues multi-entorno sin esquemas rígidos. |
| Frontend           | **React 19 + react-scripts** | React es el stack usado en TP anteriores y facilita testing con Jest/RTL. `react-scripts` mantiene configuración mínima.                                         |
| Pruebas unitarias  | **Go test + Jest**           | Nativas de cada stack. Facilitan obtener coverage (`go test -coverprofile`, `jest --coverage`).                                                                  |
| Pruebas E2E        | **Cypress 13**               | Tiene intercepts para mockear backend, integra videos/screenshots automáticamente y corre bien en GitHub Actions headless Chrome.                                |
| Análisis estático  | **SonarCloud**               | Servicio SaaS gratuito para proyectos públicos. Ofrece reglas para Go y JS, integra con GitHub Checks y Quality Gates.                                           |
| CI/CD              | **GitHub Actions**           | El repo ya está en GitHub, así que usar Actions evita configurar otro runner. Secrets (`SONAR_TOKEN`) viven en el mismo entorno.                                 |
| Orquestación local | **Docker Compose**           | Permite levantar QA/Prod separados con 6 contenedores (mongo/api/front x entorno) y políticas `depends_on` para ordenar los servicios.                           |

## 3. Arquitectura resumida

- **QA**: `mongo_db_qa (27018)`, `go_api_qa (8081)`, `frontend_qa (3000)`.
- **PROD**: `mongo_db_prod (27019)`, `go_api_prod (8082)`, `frontend_prod (3001)`.
- Cada servicio tiene su volumen (`mongo_data_*`) para persistencia y se comunica vía red interna de Compose. Esto permite probar features en QA sin tocar datos de producción.

## 4. Estrategia de calidad

### 4.1 Coverage

- **Backend**: `go test ./... -coverprofile=coverage.out` genera Cobertura en formato compatible con Sonar (`backend/coverage.out`).
- **Frontend**: Jest corre con `--coverage --watchAll=false` y guarda `frontend/coverage/lcov.info`. Se definió un umbral global del 70 %. Al subir lógica de edición de Todos la cobertura subió a ~85 % siguiendo el reporte mostrado por `npm test -- --coverage`.
- Ambos reportes se suben como artifacts (`backend-coverage`, `frontend-coverage-report`) para que SonarCloud los lea posteriormente.

### 4.2 Análisis estático

- `sonar-project.properties` referencia `backend` y `frontend/src`, excluye directorios `coverage`, `node_modules` y tests.
- SonarCloud se ejecuta con `SonarSource/sonarcloud-github-action@v2`. Se pasa `projectBaseDir` al workspace y `sonar.branch.name=${{ github.ref_name }}` para que reconozca PRs / branch main.
- Quality Gate se refuerza con `SonarSource/sonarqube-quality-gate-action@v1.1.0`. Si SonarCloud marca rojo (bugs críticos, coverage nuevo < 70 %, duplicación alta, etc.) el job falla y no se llega al summary.

### 4.3 Pruebas E2E

- Cypress se configuró con `cypress.config.js` (baseUrl parametrizable vía `CYPRESS_BASE_URL`).
- Specs (`frontend/cypress/e2e/todos.cy.js`) cubren tres flujos:
  1. Crear tarea → se intercepta `POST /todos` y se valida la UI.
  2. Editar tarea → se intercepta `PUT /todos/:id` y se usan nuevos `data-cy` agregados al frontend (`todo-input`, `edit-<id>`, `save-<id>`).
  3. Manejo de error → mockea `500` y verifica el toast.
- El script `npm run e2e` usa `start-server-and-test` para levantar `npm start`, esperar `http://localhost:3000` y correr Cypress en Chrome headless. Videos y screenshots se adjuntan como artifact `cypress-artifacts` aunque la suite falle.

## 5. Pipeline Full CI (GitHub Actions)

```text
jobs:
  backend-tests       → go test + coverage + artifact
  frontend-tests      → npm ci + Jest coverage + artifact
  sonarcloud-analysis → necesita ambos, descarga artifacts, sonar scan + quality gate
  e2e-tests           → depende de backend/frontend, corre npm run e2e + sube videos
  summary             → depende de todos; imprime estado final solo si pasaron o si e2e falló (para investigar artifacts)
```

Consideraciones clave:

- `actions/setup-go@v5` apunta al `go.mod` para evitar hardcodear versiones.
- Node 22.x se cachea por npm lockfile.
- Los permissions se dejaron en `contents/pull-requests: read` para permitir comentarios de Sonar en PRs.
- El job `sonarcloud-analysis` se limita a `main` o PRs contra `main` para no contaminar ramas efímeras.
- El summary imprime check-list textual (`Backend OK`, `Frontend OK`, `Análisis SonarCloud`, `E2E ejecutado`).

## 6. Documentación y defensa

- Este archivo (`Decisiones.md`) actúa como bitácora: explica stack, razones técnicas y cómo se garantizan métricas.
- Para la entrega formal:
  - Capturas del dashboard de SonarCloud con Quality Gate `Passed`.
  - Reporte de coverage (CLI de Jest y `go tool cover -func=coverage.out`).
  - Logs de Actions mostrando artifacts subidos.
  - Explicación de los flujos Cypress (adjuntar video/screenshot si es necesario).

## 7. Conclusiones personales

1. **Stack poliglota no es problema si se automatiza**: tener Go + React + Mongo es viable siempre que se consolide en un pipeline reproducible; Compose ayudó a aislar QA/Prod.
2. **Quality Gate real**: no basta con correr Sonar; bloquear el pipeline evita merges que rompan métricas.
3. **E2E derivan cambios en el código**: para satisfacer Cypress se añadieron `data-cy` y lógica de edición real. Esto elevó también el coverage de las unitarias.
4. **Artifacts importan**: guardar coverage y videos permite auditar fallas sin re-ejecutar todo.
5. **Próximos pasos**: agregar despliegue automático luego de pasar el gate, y publicar reportes (coverage/sonar) como badges en el README para mantener la visibilidad del estado del proyecto.
