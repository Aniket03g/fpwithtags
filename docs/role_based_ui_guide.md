# Role-Based UI Rendering Guide

This guide documents the patterns and implementation details for role-based UI rendering in FeaturePlus HTMX frontend templates.

## Overview

FeaturePlus supports two primary user roles:
- **Manager**: Can approve PRs, manage releases, and oversee team activities
- **Developer**: Can submit PRs, mark PRs as tested, and work on assigned tasks

The UI adapts to show appropriate content and actions based on the user's role without requiring separate templates for each role.

## Implementation Pattern

### Role Check Syntax

All templates use the following consistent pattern for role-based conditional rendering:

```html
{{ if eq .CurrentUser.Role "manager" }}
  <!-- Manager-specific content -->
{{ else }}
  <!-- Developer-specific content -->
{{ end }}
```

For more complex conditions:

```html
{{ if and (eq .CurrentUser.Role "manager") (ne .PR.Status "Approved") }}
  <!-- Content only for managers when PR is not approved -->
{{ end }}
```

### Common Use Cases

1. **Navigation Items**
   - Managers see additional links for "Releases" and "Team Management"
   - Both roles see common links (Projects, Features, Tasks, PRs)

2. **Dashboard Sections**
   - Managers see "PRs pending approval", "Release status", "Team overview"
   - Developers see "My assigned tasks", "My PRs", "Feature progress"

3. **Action Buttons**
   - Managers see "Approve PR" buttons
   - Developers see "Mark as Tested" buttons
   - Both see common actions like "View on GitHub"

4. **Form Fields**
   - Managers see additional fields for priority and release assignment
   - Both roles see common fields like title and description

## Template Examples

### Navigation Template

```html
<ul class="nav-menu">
  <!-- Common items for all roles -->
  <li><a href="/dashboard">Dashboard</a></li>
  <li><a href="/projects">Projects</a></li>
  
  <!-- Manager-only items -->
  {{ if eq .CurrentUser.Role "manager" }}
    <li><a href="/releases">Releases</a></li>
    <li><a href="/team">Team Management</a></li>
  {{ end }}
</ul>
```

### PR Actions Template

```html
<div class="actions">
  {{ if eq .CurrentUser.Role "manager" }}
    {{ if ne .PR.Status "Approved" }}
      <button hx-post="/prs/{{.PR.ID}}/approve">
        Approve PR
      </button>
    {{ end }}
  {{ else }}
    {{ if not .PR.Tested }}
      <button hx-post="/prs/{{.PR.ID}}/test">
        Mark as Tested
      </button>
    {{ end }}
  {{ end }}
</div>
```

## Templates Using Role-Based Rendering

The following templates have been updated to include role-based UI rendering:

1. **Navigation and Layout**
   - `dashboard.html` - Main navigation sidebar with role-specific menu items

2. **Dashboard Views**
   - `dashboard-status.html` - Dashboard content with role-specific sections

3. **PR Templates**
   - `_pr_row.html` - PR list row with role-specific action buttons
   - `pr-detail.html` - Detailed PR view with role-specific sections and actions

4. **Release Management (Manager Only)**
   - `release-list.html` - List of releases with management actions
   - `release-detail.html` - Detailed release view with management options
   - `release-form.html` - Form for creating/editing releases

5. **Team Management (Manager Only)**
   - `team-management.html` - Team overview and management interface

## Best Practices

1. **Keep Templates DRY**
   - Use conditional blocks within existing templates rather than duplicating entire templates
   - Extract common patterns to partial templates when appropriate

2. **Consistent Role Checks**
   - Always use `.CurrentUser.Role` for role checks
   - Use the exact string comparison: `eq .CurrentUser.Role "manager"`

3. **Combine with State Conditions**
   - Combine role checks with state conditions for more precise control
   - Example: `{{ if and (eq .CurrentUser.Role "manager") (ne .PR.Status "Approved") }}`

4. **Maintain Consistent Styling**
   - Keep styling and layout consistent with existing design
   - Use the same CSS classes for similar elements across roles

5. **Use HTMX for Dynamic Loading**
   - Use HTMX for dynamic fragment loading to keep the UI responsive
   - Load role-specific content asynchronously when possible

## Reference Examples

For more examples and patterns, see the `role_based_rendering_examples.html` template in the templates directory.
