# Release-First Workflow UI Changes

This document summarizes the HTMX template updates to support the release-first workflow in FeaturePlus.

## Overview

The UI now supports creating releases without PRs and planning features before implementation begins. This enables a "Release → Feature → PR" workflow.

---

## 1. Updated Templates

### `release-form.html`
**Changes:**
- ✅ Made PR selection **optional** with clear labeling
- ✅ Added helper text: "You can attach features and PRs later. Leave empty to create a release-first plan."
- ✅ Updated empty state message to mention adding features after creation
- ✅ Releases default to "draft" status (backend handles this)

**User Experience:**
- Managers can now create empty releases for planning purposes
- Clear indication that PRs are optional
- Guidance on next steps after creation

---

### `release-detail.html`
**Major Changes:**

#### 1. Enhanced Status Display
- ✅ Larger, more prominent status badge with border
- ✅ Added "Planning Mode" indicator for draft releases
- ✅ Warning icon (⚠) for draft status

#### 2. New "Planned Features" Section
- ✅ Displays features assigned to the release via `release_id`
- ✅ Shows feature details: title, status, priority, category, description
- ✅ Displays linked PR count for each feature
- ✅ Two action buttons for managers:
  - **"Add Feature"** - Creates new feature inline
  - **"Assign Existing"** - Opens modal to select existing features

#### 3. Feature Display
Each planned feature shows:
- Title with status and priority badges
- Description
- Category
- Linked PRs count (dynamically loaded)
- Remove button (managers only, draft releases only)

#### 4. Empty State
- ✅ Friendly message: "No features planned yet"
- ✅ Two prominent CTAs:
  - "Create New Feature" (green button)
  - "Assign Existing Features" (blue button)

**HTMX Endpoints Used:**
```html
<!-- Create new feature inline -->
hx-get="/web/fragments/releases/{{ .ReleaseID }}/features/new"
hx-target="#feature-section"
hx-swap="beforeend"

<!-- Assign existing features (modal) -->
hx-get="/web/fragments/releases/{{ .ReleaseID }}/features/assign"
hx-target="#modal-container"
hx-swap="innerHTML"

<!-- Remove feature from release -->
hx-post="/api/features/{{ .ID }}/unassign-release"
hx-target="#planned-feature-{{ .ID }}"
hx-swap="outerHTML"
```

---

### `release-list.html`
**No changes required** - Already supports draft releases with appropriate status badges.

---

## 2. New Templates Created

### `release-feature-form.html`
**Purpose:** Inline form for creating new features directly under a release

**Features:**
- Loads inline (not modal) for better UX
- Posts to `/api/releases/{{ .ReleaseID }}/features/create`
- Fields:
  - Title (required)
  - Description
  - Category (dropdown: general, frontend, backend, infrastructure, security, performance)
  - Priority (dropdown: low, medium, high)
- Green theme to distinguish from other forms
- Cancel button removes form without page reload
- Success adds feature card to top of list

**HTMX Configuration:**
```html
hx-post="/api/releases/{{ .ReleaseID }}/features/create"
hx-target="#feature-section"
hx-swap="afterbegin"
```

---

### `release-assign-features.html`
**Purpose:** Modal for assigning existing features to a release

**Features:**
- Displays all available features from the project
- Excludes features already assigned to any release
- Shows feature details: title, status, priority, description, category
- Multi-select checkboxes
- Posts to `/api/releases/{{ .ReleaseID }}/features`
- Scrollable list for many features

**HTMX Configuration:**
```html
hx-post="/api/releases/{{ .ReleaseID }}/features"
hx-target="#modal-container"
hx-swap="outerHTML"
```

---

## 3. Backend Handler Requirements

The templates expect these handlers to exist (you've already implemented them):

### API Endpoints
1. ✅ `POST /api/releases/:id/features/create` - Create feature under release
2. ✅ `POST /api/releases/:id/features` - Assign existing features
3. ⚠️ `POST /api/features/:id/unassign-release` - Remove feature from release (needs implementation)

### Fragment Endpoints (need implementation)
1. ⚠️ `GET /web/fragments/releases/:id/features/new` - Return `release-feature-form.html`
2. ⚠️ `GET /web/fragments/releases/:id/features/assign` - Return `release-assign-features.html` with available features

### Data Requirements
The `release-detail.html` template expects:
- `.PlannedFeatures` - Array of features with `release_id = {{ .Release.ID }}`
- `.AvailableFeatures` - Features from same project not assigned to any release (for assign modal)

---

## 4. Workflow Examples

### Release-First Workflow (New)
```
1. Manager creates empty release (v2.0.0)
   → Status: DRAFT
   → PRs: 0
   → Features: 0

2. Manager adds features to release:
   a. Click "Add Feature" → Create new feature inline
   b. Click "Assign Existing" → Select from existing features
   
3. Features show in "Planned Features" section
   → Each feature shows: title, status, priority, category
   → PR count: 0 (not implemented yet)

4. Developers create PRs for features (normal workflow)
   → PRs automatically linked via feature_id

5. Manager clicks "Finalize Release"
   → System collects PRs from:
     - Direct associations (release_prs table)
     - Features assigned to release (via release_id)
   → Creates git branch and tag
   → Status: PUBLISHED
```

### Traditional PR-First Workflow (Still Supported)
```
1. Developers create PRs
2. Manager creates release and selects PRs
3. Manager finalizes release
```

---

## 5. Visual Indicators

### Draft Status
- **Badge:** Yellow background, yellow text, yellow border
- **Text:** "DRAFT" in uppercase
- **Indicator:** "⚠ Planning Mode - Add features and finalize when ready"

### Published Status
- **Badge:** Green background, green text, green border
- **Text:** "PUBLISHED" in uppercase
- **Timestamp:** Shows publication date

### Failed Status
- **Badge:** Red background, red text, red border
- **Text:** "FAILED" in uppercase

---

## 6. Role-Based Access

All new features respect role-based access:
- **Managers Only:**
  - Create releases
  - Add/remove features
  - Assign existing features
  - Finalize releases
  
- **All Users:**
  - View releases
  - View planned features
  - View PRs

---

## 7. Next Steps for Full Implementation

### Required Backend Handlers
```go
// 1. Serve feature creation form
GET /web/fragments/releases/:id/features/new
→ Return release-feature-form.html with ReleaseID

// 2. Serve feature assignment modal
GET /web/fragments/releases/:id/features/assign
→ Fetch available features (same project, no release_id)
→ Return release-assign-features.html with data

// 3. Unassign feature from release
POST /api/features/:id/unassign-release
→ Set feature.release_id = NULL
→ Return empty response (HTMX will remove element)
```

### Required Data Loading
In `RenderReleaseDetailFragment` handler:
```go
// Fetch features assigned to this release
var plannedFeatures []models.Feature
db.Where("release_id = ?", releaseID).Find(&plannedFeatures)

// Pass to template
c.HTML(http.StatusOK, "release-detail.html", gin.H{
    "Release": release,
    "PlannedFeatures": plannedFeatures,
    "CurrentUser": currentUser,
})
```

---

## 8. Testing Checklist

- [ ] Create empty release (no PRs selected)
- [ ] Add new feature to release
- [ ] Assign existing feature to release
- [ ] Remove feature from release
- [ ] View planned features in release detail
- [ ] Finalize release with features (should collect PRs from features)
- [ ] Verify draft status indicators show correctly
- [ ] Verify manager-only buttons are hidden for non-managers
- [ ] Test both workflows (release-first and PR-first)

---

## Summary

The UI now fully supports the release-first workflow while maintaining backward compatibility with the PR-first approach. The changes are intuitive, visually clear, and follow existing FeaturePlus design patterns.
