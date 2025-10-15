# Import Project UI Guide

## Visual Overview

### 1. Project List Page - Before

```
┌─────────────────────────────────────────────────────────┐
│  Projects                        [+ Create Project]     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ Project  │  │ Project  │  │ Project  │             │
│  │    1     │  │    2     │  │    3     │             │
│  └──────────┘  └──────────┘  └──────────┘             │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 2. Project List Page - After (With Import Button)

```
┌─────────────────────────────────────────────────────────────────┐
│  Projects              [📥 Import Project] [+ Create Project]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │
│  │ Project  │  │ Project  │  │ Project  │                     │
│  │    1     │  │    2     │  │    3     │                     │
│  └──────────┘  └──────────┘  └──────────┘                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Button Colors:**
- 🟢 **Import Project**: Green (`bg-green-600`)
- 🔵 **Create Project**: Blue (`bg-blue-600`)

---

## 3. Import Modal

When user clicks "Import Project":

```
┌─────────────────────────────────────────────────────────┐
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  📥  Import Project from GitHub MCP            │    │
│  ├────────────────────────────────────────────────┤    │
│  │                                                 │    │
│  │  Project ID *                                   │    │
│  │  ┌─────────────────────────────────────────┐  │    │
│  │  │ e.g., github_project_demo               │  │    │
│  │  └─────────────────────────────────────────┘  │    │
│  │  Enter the JSON filename (without .json)      │    │
│  │                                                 │    │
│  │  Project Name *                                │    │
│  │  ┌─────────────────────────────────────────┐  │    │
│  │  │ e.g., My Imported Project               │  │    │
│  │  └─────────────────────────────────────────┘  │    │
│  │                                                 │    │
│  │  Description (Optional)                        │    │
│  │  ┌─────────────────────────────────────────┐  │    │
│  │  │ Add a description...                    │  │    │
│  │  │                                         │  │    │
│  │  └─────────────────────────────────────────┘  │    │
│  │                                                 │    │
│  │  ℹ️  This will import features, tasks, and    │    │
│  │     configurations from the GitHub MCP-        │    │
│  │     generated JSON file.                       │    │
│  │                                                 │    │
│  │              [Cancel]  [Import Project]        │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 4. Loading State

When form is submitted:

```
┌─────────────────────────────────────────────────────────┐
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │  📥  Import Project from GitHub MCP            │    │
│  ├────────────────────────────────────────────────┤    │
│  │                                                 │    │
│  │  [Form fields shown above]                     │    │
│  │                                                 │    │
│  │              [Cancel]  [⏳ Import Project]     │    │
│  │                         ↑                       │    │
│  │                    Spinner appears              │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 5. Success State

After successful import:

```
┌─────────────────────────────────────────────────────────────────┐
│                    ┌──────────────────────────────────────┐     │
│                    │ ✅ Project "My Imported Project"     │     │
│                    │    imported successfully!            │     │
│                    └──────────────────────────────────────┘     │
│  Projects              [📥 Import Project] [+ Create Project]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐                 │
│  │ Project  │  │ Project  │  │ My Imported  │ ← Green border  │
│  │    1     │  │    2     │  │   Project    │   + pulse       │
│  └──────────┘  └──────────┘  └──────────────┘                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Success Indicators:**
1. ✅ Toast notification (top-right, 3 seconds)
2. 🟢 Green border on new project
3. ✨ Pulse animation
4. 🚫 Modal automatically closed

---

## User Interaction Flow

### Step-by-Step

1. **Navigate to Projects**
   ```
   User → Dashboard → Projects
   ```

2. **Click Import Button**
   ```
   Click: [📥 Import Project]
   Result: Modal appears
   ```

3. **Fill Form**
   ```
   Project ID: github_project_demo
   Project Name: My Test Import
   Description: Testing the import feature
   ```

4. **Submit**
   ```
   Click: [Import Project]
   Result: Spinner appears
   ```

5. **Wait for Import**
   ```
   Backend: Reads JSON → Creates project → Creates features/tasks
   Time: ~500ms
   ```

6. **See Results**
   ```
   Modal: Closes automatically
   Toast: "✅ Project imported successfully!"
   List: New project appears with green border
   ```

---

## Form Validation

### Required Fields
- ❌ Empty project_id → "Please fill out this field"
- ❌ Empty project_name → "Please fill out this field"
- ✅ Both filled → Form can submit

### Real-time Feedback
- Placeholder text shows examples
- Helper text explains each field
- Info box explains the import process

---

## Error States

### 1. Template Not Found (404)

```
┌─────────────────────────────────────────────────────────┐
│  ┌────────────────────────────────────────────────┐    │
│  │  📥  Import Project from GitHub MCP            │    │
│  ├────────────────────────────────────────────────┤    │
│  │                                                 │    │
│  │  ⚠️ Error: Import template not found           │    │
│  │     Please check the project ID and try again  │    │
│  │                                                 │    │
│  │  [Form remains open for correction]            │    │
│  └────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### 2. Authentication Required (401)

```
Redirects to login page
```

### 3. Server Error (500)

```
┌─────────────────────────────────────────────────────────┐
│  ⚠️ Error: Failed to import project                     │
│     Please try again or contact support                 │
└─────────────────────────────────────────────────────────┘
```

---

## Keyboard Shortcuts

- `Tab` - Navigate between fields
- `Enter` - Submit form (when in input field)
- `Esc` - Close modal
- Click outside modal - Close modal

---

## Mobile Responsive

### Desktop (> 768px)
```
[Import Project] [Create Project]  ← Side by side
```

### Mobile (< 768px)
```
[Import Project]  ← Stacked
[Create Project]
```

Modal adapts to screen size with proper padding and scrolling.

---

## Accessibility Features

1. **ARIA Labels**
   - `aria-modal="true"` on modal
   - `role="dialog"` for screen readers

2. **Focus Management**
   - First field auto-focused when modal opens
   - Focus trapped within modal
   - Focus restored when modal closes

3. **Visual Indicators**
   - Required fields marked with `*`
   - Color contrast meets WCAG AA standards
   - Loading states clearly indicated

4. **Keyboard Navigation**
   - All interactive elements keyboard accessible
   - Logical tab order
   - Escape key closes modal

---

## Color Scheme

### Import Theme (Green)
- **Button**: `bg-green-600` → `hover:bg-green-700`
- **Icon**: Green cloud upload
- **Focus**: `ring-green-500`
- **Success**: Green toast notification

### Create Theme (Blue)
- **Button**: `bg-blue-600` → `hover:bg-blue-700`
- **Icon**: Blue plus sign
- **Focus**: `ring-blue-500`

### Neutral Elements
- **Modal Background**: White
- **Overlay**: Gray with 75% opacity
- **Text**: Gray-700 (labels), Gray-900 (headings)
- **Borders**: Gray-300

---

## Animation Details

1. **Modal Entrance**
   - Fade in overlay (300ms)
   - Scale up modal (300ms)

2. **Loading Spinner**
   - Rotate animation (infinite)
   - Appears next to button text

3. **Success Toast**
   - Slide in from right (200ms)
   - Fade out after 3 seconds (300ms)

4. **New Project Highlight**
   - Green border
   - Pulse animation (2 seconds)
   - Fades to normal state

---

## Tips for Users

1. **Finding Project ID**
   - Look in `/backend/data/imports/` folder
   - Use filename without `.json` extension
   - Example: `github_project_demo.json` → `github_project_demo`

2. **Naming Projects**
   - Use descriptive names
   - Can be different from JSON filename
   - Example: `github_project_demo` → "My GitHub Demo Project"

3. **Troubleshooting**
   - Check server logs for detailed errors
   - Verify JSON file exists and is valid
   - Ensure you're logged in
   - Try refreshing the page

---

## Comparison: Import vs Create

| Feature | Import | Create |
|---------|--------|--------|
| **Color** | Green | Blue |
| **Icon** | Cloud upload | Plus sign |
| **Speed** | Fast (reads JSON) | Slower (manual entry) |
| **Features** | Pre-defined | Manual |
| **Tasks** | Pre-defined | Manual |
| **Use Case** | GitHub repos | Custom projects |

---

## Next Steps After Import

1. **View Project**
   - Click "View" button on imported project
   - See all features and tasks

2. **Customize**
   - Edit features/tasks as needed
   - Add more features
   - Assign team members

3. **Start Working**
   - Begin tracking progress
   - Create PRs
   - Plan releases

---

This UI provides a seamless, intuitive way to import GitHub MCP-generated projects into FeaturePlus!
