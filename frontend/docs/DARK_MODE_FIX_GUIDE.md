# Dark Mode Color Fix Guide

## ✅ Completed Updates

### 1. Core Color Variables (`src/styles/colors.scss`)
- ✅ Added comprehensive CSS variables for both light and dark modes
- ✅ Added `--bg-primary` and `--bg-secondary` variables for consistent backgrounds
- ✅ Fixed `--bardhart-labletext` to use proper light color (#e5e5e5) in dark mode
- ✅ Ensured all color variables have dark mode equivalents

### 2. Global Styles (`src/styles.scss`)
- ✅ Updated body background to use `var(--bg-primary)`
- ✅ Added smooth transitions for theme changes

### 3. Shared Components
#### Sidebar (`src/app/shared/components/sidebar/`)
- ✅ Removed hardcoded `color: black !important` (changed to `var(--sidebar-text)`)
- ✅ Fixed `.bus-staff-label` to use `var(--sidebar-text)`
- ✅ Fixed extra closing brace syntax error

#### Notification Panel (`src/app/shared/components/notification-panel/`)
- ✅ Already using CSS variables properly
- ✅ Uses `var(--color-background-primary)`, `var(--color-text-primary)`, etc.

### 4. Page Components - Backgrounds Updated
- ✅ `dashboard.component.scss` - Changed to `var(--bg-primary)` and `var(--bg-secondary)`
- ✅ `bus-management.component.scss` - Updated backgrounds and stat cards
- ✅ `bus-booking.component.scss` - Changed to `var(--bg-primary)` and `var(--bg-secondary)`
- ✅ `driver-management.component.scss` - Updated to `var(--bg-primary)`
- ✅ `conductor-management.component.scss` - Updated to `var(--bg-primary)`
- ✅ `passenger-management.component.scss` - Updated to `var(--bg-primary)`
- ✅ `lounges-management.component.scss` - Updated to `var(--bg-primary)`
- ✅ `lounge-booking.component.scss` - Updated to `var(--bg-primary)`
- ✅ `notification-details.component.scss` - Already using `var(--bg-primary)` and `var(--bg-secondary)`

## 🔄 Remaining Work

### Search and Replace Pattern
To fix all remaining hardcoded colors, use the following find/replace patterns:

#### 1. Replace hardcoded white backgrounds
```scss
// Find:
background: #fff;

// Replace with:
background: var(--bg-secondary);
```

#### 2. Replace hardcoded white backgrounds (alternate)
```scss
// Find:
background: #ffffff;

// Replace with:
background: var(--bg-secondary);
```

#### 3. Fix hardcoded border colors
```scss
// Find:
border: 1px solid #e5e7eb

// Replace with:
border: 1px solid var(--color-border-light)
```

#### 4. Fix hardcoded text colors
```scss
// Find:
color: #111827

// Replace with:
color: var(--color-text-primary)
```

```scss
// Find:
color: #333333

// Replace with:
color: var(--color-text-primary)
```

### Files That Still Need Updates

#### Driver Management
- ✅ `driver-management.component.scss` - Main page updated
- 🔄 `add-driver.component.scss` - Line 90: `background: #ffffff`
- 🔄 `edit-driver.component.scss` - Line 90: `background: #ffffff`

#### Conductor Management
- ✅ `conductor-management.component.scss` - Main page updated
- 🔄 `add-conductor.component.scss` - Line 63: `background: #fff`
- 🔄 `edit-conductor.component.scss` - Line 69: `background: #fff`

#### Lounge Management
- ✅ `lounges-management.component.scss` - Main page updated
- 🔄 `view-lounge.component.scss` - Lines 50, 89, 123: `background: #fff`
- 🔄 `edit-lounge.component.scss` - Line 86: `background: #fff`
- 🔄 `add-lounge.component.scss` - Line 95: `background: #fff`

#### Lounge Booking
- ✅ `lounge-booking.component.scss` - Main page updated
- 🔄 `edit-lounge-booking.component.scss` - Line 27: `background: #fff`

#### Bus Management
- ✅ `bus-management.component.scss` - Main page updated with most colors
- 🔄 `edit-bus.component.scss` - Line 75: `background: #ffffff`

## 📋 CSS Variable Reference

### Background Colors
```scss
// Light Mode
--bg-primary: #f7f7f7;        // Page backgrounds
--bg-secondary: #ffffff;       // Card/component backgrounds

// Dark Mode
--bg-primary: #1a1a1a;        // Page backgrounds
--bg-secondary: #2d2d2d;       // Card/component backgrounds
```

### Text Colors
```scss
// Light Mode
--color-text-primary: #333333;
--color-text-secondary: #666666;

// Dark Mode
--color-text-primary: #e5e5e5;
--color-text-secondary: #b0b0b0;
```

### Border Colors
```scss
// Light Mode
--color-border-light: #e5e7eb;
--border-color: #e0e0e0;

// Dark Mode
--color-border-light: #3d3d3d;
--border-color: #2d2d2d;
```

### Button Colors
```scss
// Light Mode
--active: #0046FF;
--inactive: #FAA533;

// Dark Mode
--active: #5A8EFF;
--inactive: #FAA533;
```

## 🎨 Best Practices

1. **Always use CSS variables** instead of hardcoded colors
2. **Use semantic variable names**:
   - `--bg-primary` for page backgrounds
   - `--bg-secondary` for card/content backgrounds
   - `--color-text-primary` for main text
   - `--color-text-secondary` for secondary text
3. **Avoid `!important`** unless absolutely necessary
4. **Test in both light and dark modes** after making changes

## 🚀 Quick Fix Script

Run this in VS Code's Find & Replace (Ctrl+Shift+H):

1. **Enable Regex** (Alt+R)
2. **Set Files to Include**: `**/*.component.scss`
3. **Run these replacements in order**:

```
Find: background:\s*#fff(?:fff)?;
Replace: background: var(--bg-secondary);

Find: background:\s*#f[0-9a-f]{5,6};
Replace: background: var(--bg-primary);

Find: color:\s*#111827;
Replace: color: var(--color-text-primary);

Find: border:\s*1px\s*solid\s*#e5e7eb;
Replace: border: 1px solid var(--color-border-light);
```

## ✅ Testing Checklist

After applying fixes, test these scenarios:

- [ ] Toggle dark mode from sidebar
- [ ] Check all page backgrounds (dashboard, bus, driver, conductor, passenger, lounge)
- [ ] Verify card/table backgrounds are visible
- [ ] Test text readability in both modes
- [ ] Check notification panel appearance
- [ ] Verify sidebar styling in both modes
- [ ] Test button hover states
- [ ] Check form input styling

## 🔍 Common Issues & Solutions

### Issue: White backgrounds in dark mode
**Solution**: Replace `background: #fff` with `background: var(--bg-secondary)`

### Issue: Text not readable in dark mode
**Solution**: Replace hardcoded text colors with `var(--color-text-primary)` or `var(--color-text-secondary)`

### Issue: Borders invisible in dark mode
**Solution**: Replace hardcoded border colors with `var(--color-border-light)`

### Issue: Charts/graphs not visible
**Solution**: Ensure chart libraries use theme-aware colors from CSS variables
