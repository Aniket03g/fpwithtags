# PR Release Management Implementation

This document outlines the implementation of the PR release management functionality in FeaturePlus, which enables managers to create and finalize releases directly from the PR management screen.

## Overview

The PR release management feature allows managers to:
1. Select PRs from the PR list and create a new release
2. Fill in release details (version tag and notes) in a modal form
3. Submit the form to create a draft release
4. View the list of releases with their status
5. Finalize draft releases with a single click

All interactions are implemented using HTMX for seamless, server-rendered UI updates without full page reloads.

## Components

### 1. PR Selection and Release Creation

- **PR List UI**: The task-card.html template includes checkboxes for PR selection that appear when the "Create Release" button is clicked.
- **Continue Button**: Appears when in selection mode, collects selected PR IDs, and opens the release creation modal.
- **Release Modal**: A form for entering version tag and release notes, with hidden fields for selected PR IDs.

### 2. Release Creation Flow

- **Modal Form**: Submits JSON data to the `/releases` endpoint using HTMX.
- **CreateRelease Handler**: Validates input, creates a draft release, and returns success response.
- **UI Update**: The release list is updated dynamically when a new release is created.

### 3. Release Finalization

- **Finalize Button**: Available only for draft releases and only to managers.
- **FinalizeRelease Handler**: Validates the release, finalizes it using the shared package, updates its status, and returns the updated release row HTML.
- **UI Update**: The release row is dynamically updated to reflect the published status.

## Technical Implementation

### Templates

1. **release-modal.html**: Modal form for creating a new release
2. **_release_row.html**: Partial template for rendering a single release row
3. **release-list.html**: Main template for displaying the list of releases
4. **task-card.html**: Updated to include PR selection functionality

### Handlers

1. **WebReleaseHandler**: Serves the release creation modal with selected PR IDs
2. **ReleaseHandler**: Handles release creation and finalization

### Routes

1. `/web/fragments/release-modal`: Serves the release creation modal
2. `/releases`: Endpoint for creating a new release
3. `/releases/{id}/finalize`: Endpoint for finalizing a draft release

## HTMX Integration

The implementation uses HTMX for all dynamic UI updates:

1. **Modal Loading**: `hx-get` to load the release modal
2. **Form Submission**: `hx-post` with JSON data to create a release
3. **Release Finalization**: `hx-post` to finalize a release and swap the updated row
4. **Target Swapping**: `hx-target` and `hx-swap` to update specific parts of the UI

## Role-Based Access Control

All release management functionality is restricted to users with manager role:

1. Create Release button is only visible to managers
2. Finalize button is only visible to managers
3. Server-side validation ensures only managers can perform these actions

## Error Handling

1. Client-side validation for required fields
2. Server-side validation for input data
3. Proper error responses for invalid requests
4. Confirmation dialogs for destructive actions

## Future Enhancements

1. Add release editing functionality
2. Implement release rollback
3. Add release notes templates
4. Integrate with CI/CD pipelines for automated deployments
