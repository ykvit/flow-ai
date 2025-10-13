
# How to Contribute to Flow-AI

Thank you for your interest in contributing to Flow-AI! 🚀
We welcome contributions of all kinds: bug fixes, new features, documentation updates, and infrastructure improvements.

This document describes how we work, coding standards, and expectations for contributors.


## 🔄 Development Process (GitHub Flow)

We use the **GitHub Flow** branching model:

1. The `main` branch is always stable and deployable.
2. Create a **temporary branch** from `main` for your work.
3. When ready, open a **Pull Request** (PR) to merge into `main`.
4. Automated CI checks will run, and a maintainer will review your PR.
5. After approval and passing checks, the PR is merged and the branch deleted.

## 🌿 Branch Naming Conventions

* **Features:** `feature/short-description`
  *e.g.* `feature/user-login-form`
* **Bug Fixes:** `fix/short-description`
  *e.g.* `fix/chat-message-ordering`
* **Documentation:** `docs/short-description`
  *e.g.* `docs/update-api-examples`
* **Chores (CI/CD, refactoring, infra):** `chore/short-description`
  *e.g.* `chore/add-ci-pipeline`

## ✍️ Commit Message Guidelines

We encourage clear and consistent commit messages.
Following the [Conventional Commits](https://www.conventionalcommits.org/) format is recommended:

```
type(scope): short description
```

Examples:

* `fix(auth): prevent null token crash`
* `docs(api): update examples in Swagger`
* `chore(ci): add Trivy security scan`

## ✅ Local Development & Testing

Before pushing your branch, please run local checks to ensure consistency:

```sh
# Format code
make format

# Run lint checks
make lint

# Run backend tests
make test-backend
```

Additional commands:

* `make dev` → run development environment
* `make swag` → regenerate Swagger documentation after API changes

## 📚 Documentation References

To avoid duplication, please refer to:

* **Project overview & setup:** [README.md](./README.md)
* **Architecture & design decisions:** [DOCUMENTATION.md](./DOCUMENTATION.md)
* **API details & Swagger guide:** [API.md](./API.md)
* **Backend-specific setup:** [backend/README.md](./backend/README.md)
* **Infrastructure & Docker setup:** [docker/README.md](./docker/README.md)

