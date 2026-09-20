# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-09-20

### Added

- **Download a folder as a ZIP file.** Folders can now be downloaded in one go
  from the folder action menu via a new **Download as ZIP** entry, removing the
  need to download each object individually.
  - Every object below the folder prefix is listed recursively and written into
    a single archive that preserves the internal folder structure; entry names
    are relative to the downloaded folder.
  - The archive is streamed to the browser as it is built, so no temporary file
    is created and memory usage is independent of the folder size.
  - Zero-byte folder marker objects are skipped. A folder containing no objects
    is rejected with `HTTP 404` instead of producing an empty archive.
  - The archive is named after the folder (`holiday.zip`); downloading a bucket
    root produces `<bucket>.zip`. The `Content-Disposition` header carries both
    an ASCII fallback and a UTF-8 encoded file name.
  - New endpoint: `GET /api/buckets/{bucket}/objects/{folder}/zip`.

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

[1.4.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.4.0
[1.3.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.3.0
[1.2.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.2.0
[1.0.0]: https://github.com/dimuthnc/s3manager/releases/tag/v1.0.0
