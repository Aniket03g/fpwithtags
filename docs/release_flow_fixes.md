# Release Flow Fixes

## Issues Fixed

1. **Incorrect API Endpoint in Release Modal Form**
   - The release modal form was posting to `/releases` instead of the correct API endpoint `/api/releases`
   - Fixed by updating the `hx-post` attribute in `release-modal.html`

2. **Incomplete Web Release Handler**
   - The `WebReleaseHandler` was missing methods for rendering release lists and details
   - Added methods for `RenderReleasesList`, `RenderReleaseDetail`, and `RenderReleaseRow`
   - Updated constructor to include PR repository dependency

3. **Incorrect API Endpoints in Release Row Template**
   - The finalize and delete buttons in `_release_row.html` were using incorrect endpoints
   - Updated to use `/api/releases/:id/finalize` and `/api/releases/:id` respectively

4. **Missing Web Routes for Releases**
   - Added proper web routes for releases in `routes/release_routes.go`
   - Added fragment route for the release modal

## Testing Instructions

1. Navigate to the PR list page
2. Select one or more PRs and click "Create Release"
3. Fill in the release form and submit
4. Verify the release is created successfully
5. Test finalizing a release from the releases list
6. Verify the release status changes to "Released"

## Technical Details

The release flow now follows the same pattern as the PR flow:
- API endpoints under `/api/releases` for data operations
- Web routes under `/web/releases` for HTML rendering
- Fragment routes under `/web/fragments` for HTMX partial updates
- Proper HTMX attributes for form submission and DOM updates
