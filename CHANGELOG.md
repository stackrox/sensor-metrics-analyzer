# Changelog

All notable changes to this project are documented in this file.

## 0.0.5

- Split histogram `+Inf` overflow reporting by label-set so problematic series are listed individually.
- Added metric description (Prometheus HELP text) to overflow findings.
- Improved overflow wording and added on-the-fly unit guessing for histogram outputs.
- Improved release/docs workflow (release target, GitHub Pages docs, container registry update).

## 0.0.4

- Added containerization and Helm chart support for deployment.
- Fixed Helm chart packaging/versioning issues for `0.0.4`.
- Continued rule quality improvements and review metadata propagation.

## 0.0.3

- Introduced rule review metadata and broader rule/test polishing.
- Added/expanded tests for gauge-threshold rules.
- Improved readability of numeric output formatting in reports.

## 0.0.2

- Added web server mode for serving analyzer output.

## 0.0.1

- Initial usable release with the results of the AI hackathon, including:
  - interactive TUI mode,
  - demo assets,
  - generic histogram rule support,
  - first web-server iteration.
