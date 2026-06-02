# 🎨 Color System Reference

## Overview
This document provides a clear reference for all CSS color variables used in the STS Admin Web App. The color system is organized into logical categories for easy maintenance.

---

## 📋 Color Categories

### 1. **Primary Brand Colors**
Core colors that define your brand identity.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--primary-color` | `#647FBC` | `#8BA3D7` | Primary brand color |
| `--secondary-color` | `#91ADC8` | `#A8C5D1` | Secondary brand color |
| `--tertiary-color` | `#AED6CF` | `#C2E5DD` | Tertiary brand color |

### 2. **Page Backgrounds**
Main background colors for pages and components.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--bg-primary` | `#f7f7f7` | `#1a1a1a` | Main page background |
| `--bg-secondary` | `#ffffff` | `#2d2d2d` | Cards, tables, content areas |

### 3. **Text Colors**
Typography colors for different text hierarchies.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--text-primary` | `#333333` | `#e5e5e5` | Main headings, important text |
| `--text-secondary` | `#666666` | `#b0b0b0` | Descriptions, labels |
| `--color-text-muted` | `#0046FF` | `#7BA3FF` | Links, accents, highlights |
| `--color-text-inverse` | `#ffffff` | `#1a1a1a` | Text on contrasting backgrounds |

### 4. **Borders & Shadows**
Edge and depth styling.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--border-color` | `#e0e0e0` | `#2d2d2d` | Standard borders |
| `--color-border-light` | `#e5e7eb` | `#3d3d3d` | Tables, inputs, dividers |
| `--shadow-color` | `rgba(100,127,188,0.1)` | `rgba(255,255,255,0.1)` | Box shadows |

### 5. **Buttons & Interactive Elements**
Action colors for buttons and interactive components.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--color-buttns-light` | `#0046FF` | `#5A8EFF` | Primary action buttons |
| `--color-buttns-dark` | `#a3a3a3` | `#6d6d6d` | Secondary buttons |
| `--active` | `#0046FF` | `#5A8EFF` | Active/selected states |
| `--inactive` | `#FAA533` | `#FAA533` | Inactive/warning states |

### 6. **Status Colors**
Semantic colors for success, warning, error states.

| Variable | Light & Dark Mode | Usage |
|----------|-------------------|-------|
| `--color-success-400` | `#34d399` | Success highlights |
| `--color-success-500` | `#10b981` | Success primary |
| `--color-success-600` | `#059669` | Success dark |
| `--color-warning-400` | `#fad577` | Warning highlights |
| `--color-warning-500` | `#f59e0b` | Warning primary |
| `--color-danger-500` | `#ef4444` | Error/danger states |
| `--color-error-400` | `#f87171` | Error highlights |
| `--color-info-400` | `#22d3ee` | Info highlights |
| `--color-info-500` | `#06b6d4` | Info primary |

### 7. **Charts & Data Visualization**
Colors specifically for charts and graphs.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--color-bargraph` | `#647FBC` | `#8BA3D7` | Bar chart primary |
| `--color-bargraphgradient` | `#91ADC8` | `#A8C5D1` | Bar chart gradient |
| `--color-chart-p` | `#666666` | `#b0b0b0` | Chart text/labels |
| `--chart-background` | `#ffffff` | `#2d2d2d` | Chart container |
| `--bardhart-labletext` | `#333333` | `#e5e5e5` | Bar chart labels |

### 8. **Component Specific**
Colors for specific UI components.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--statbackground` | `#ffffff` | `#2d2d2d` | Stat cards, filter backgrounds |
| `--color-table-header` | `#e5e7eb` | `#2d2d2d` | Table headers |
| `--searchbar-background` | `#ffffff` | `#2d2d2d` | Search input fields |
| `--searchborder` | `#0046FF` | `#5A8EFF` | Search focus border |
| `--color-toolbar-brdborder` | `#0046FF` | `#5A8EFF` | Toolbar borders |
| `--gray` | `#a3a3a3` | `#6d6d6d` | Utility gray |
| `--color-stat` | `#91ADC8` | `#A8C5D1` | Stat card accents |

### 9. **Sidebar**
Dedicated sidebar navigation colors.

| Variable | Light Mode | Dark Mode | Usage |
|----------|------------|-----------|-------|
| `--sidebar-background` | `#ffffff` | `#1e1e1e` | Sidebar background |
| `--sidebar-text` | `#333333` | `#e5e5e5` | Sidebar text |
| `--sidebar-hover` | `#f5f5f5` | `#2d2d2d` | Hover state |
| `--sidebar-active` | `#0046FF` | `#5A8EFF` | Active menu item |

### 10. **Neutral Scale**
Grayscale palette for various UI elements.

| Variable | Light Mode | Dark Mode |
|----------|------------|-----------|
| `--color-neutral-100` | `#f5f5f5` | `#2d2d2d` |
| `--color-neutral-200` | `#e5e5e5` | `#3d3d3d` |
| `--color-neutral-300` | `#d4d4d4` | `#4d4d4d` |
| `--color-neutral-400` | `#a3a3a3` | `#6d6d6d` |
| `--color-neutral-500` | `#737373` | `#8d8d8d` |
| `--color-neutral-600` | `#525252` | `#b0b0b0` |
| `--color-neutral-700` | `#404040` | `#d0d0d0` |
| `--color-neutral-800` | `#262626` | `#e5e5e5` |
| `--color-neutral-900` | `#171717` | `#f5f5f5` |

---

## 🎯 Usage Guidelines

### 1. **Always Use CSS Variables**
```scss
// ✅ Good
background: var(--bg-primary);
color: var(--text-primary);

// ❌ Bad
background: #f7f7f7;
color: #333333;
```

### 2. **Use Semantic Names**
Choose the variable that best describes its purpose:
- `--bg-primary` for page backgrounds
- `--bg-secondary` for card/content backgrounds
- `--text-primary` for main text
- `--text-secondary` for descriptive text

### 3. **Legacy Variables**
Some components still use legacy variable names for backward compatibility:
- `--color-text-primary` → Use `--text-primary` in new code
- `--color-background-primary` → Use `--bg-primary` in new code

### 4. **Status Colors Stay Consistent**
Success, warning, error, and info colors remain the same in both light and dark modes for consistency.

---

## 📊 Color Relationships

```
Primary Brand → Charts & Graphs
├── --primary-color (#647FBC) → --color-bargraph
├── --secondary-color (#91ADC8) → --color-bargraphgradient
└── --tertiary-color (#AED6CF) → --color-stat

Interactive Elements
├── --active (#0046FF) → --color-buttns-light
├── --active (#0046FF) → --searchborder
└── --active (#0046FF) → --sidebar-active
```

---

## 🔄 Dark Mode Behavior

### Automatic Color Switching
When `[data-theme="dark"]` is applied to the `<html>` element, all CSS variables automatically switch to their dark mode values.

### Implementation
```typescript
// In sidebar.component.ts
toggleDarkMode(): void {
  if (isPlatformBrowser(this.platformId)) {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
  }
}
```

---

## 🧪 Testing Your Colors

### Checklist
- [ ] Test all pages in light mode
- [ ] Test all pages in dark mode
- [ ] Check text contrast ratios (WCAG AA minimum)
- [ ] Verify chart colors are visible in both modes
- [ ] Test hover states on interactive elements
- [ ] Check form inputs and focus states
- [ ] Verify sidebar navigation in both modes
- [ ] Test table headers and rows
- [ ] Check button states (normal, hover, active, disabled)

---

## 🎨 Color Modifications

### Adding New Colors
1. Add to `:root` section for light mode
2. Add to `[data-theme="dark"]` section for dark mode
3. Use clear, semantic names
4. Document in this reference file
5. Group with related colors

### Modifying Existing Colors
1. Test changes in both light and dark modes
2. Check all pages that use the color
3. Verify accessibility (contrast ratios)
4. Update this documentation

---

## 📝 Notes

- **Removed Variables**: Unused variables like `--lightblu`, `--color-primary-500-rgb`, `--color-info-500-rgb`, and `--background-color` have been removed
- **Compatibility**: Legacy variable names are kept for backward compatibility with existing components
- **Organization**: Colors are grouped by purpose for easier maintenance
- **Comments**: Inline comments explain the usage of each variable

---

## 🔗 Related Files

- **Main Color File**: `src/styles/colors.scss`
- **Global Styles**: `src/styles.scss`
- **Dark Mode Implementation**: `src/app/shared/components/sidebar/sidebar.component.ts`
- **Usage Guide**: `/DARK_MODE_FIX_GUIDE.md`
