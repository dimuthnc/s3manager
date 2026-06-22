# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-06-22

### Added

- **Move and copy files/folders between bucket locations.** Objects and folders
  can now be moved or copied to another bucket/folder — including **across
  configured S3 instances** — directly from the UI, removing the previous
  download-and-re-upload workaround.
  - **Move** copies the object then deletes the source; **Copy** keeps the original.
  - Same-instance transfers use server-side `CopyObject` (no data flows through
    the app server); cross-instance transfers stream download → upload.
  - Folders are expanded recursively, preserving the folder name and internal
    structure under the destination prefix.
  - Sources are deleted only after every copy succeeds, so a mid-way failure
    never loses data. In-place no-op moves and requests missing a destination
    bucket are rejected with `HTTP 400`.
  - New **Copy to…** / **Move to…** actions in the object action menu (and a new
    action menu for folders) plus a destination modal to choose the target
    instance, bucket, and folder. **Move** is gated behind `ALLOW_DELETE`;
    **Copy** is always available.
  - New endpoint: `POST /api/buckets/{bucket}/objects/{object}/move`.

## [1.2.0]

### Changed

- UI improvements and bundled sample data for local development.

## [1.0.0]

### Added

- Initial release of the fork with a redesigned UI based on shadcn/ui design
  principles (replacing Materialize CSS), SVG icons, and improved modals,
  toasts, tables, and empty states.

[1.3.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.3.0
[1.2.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.2.0
[1.0.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.0.0
