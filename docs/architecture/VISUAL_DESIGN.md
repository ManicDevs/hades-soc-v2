# Hades Toolkit - Visual Design Specification

## Overview
The Hades Toolkit features a dark, professional security platform interface with glass-morphism effects, gradient accents, and a modern enterprise aesthetic.

## Color Scheme
- **Primary Background**: #0a0a0a (near-black)
- **Secondary Background**: #1a1a1a (dark gray)
- **Tertiary Background**: #2a2a2a (medium gray)
- **Primary Accent**: #3b82f6 (blue)
- **Success**: #22c55e (green)
- **Warning**: #f59e0b (orange)
- **Error**: #ef4444 (red)
- **Text Primary**: #f3f4f6 (white)
- **Text Muted**: #9ca3af (gray)

## Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Header (72px height)                                                            │
│ ┌─────────┐ ┌─────────────────────────────────────────────────────────────┐   │
│ │ Sidebar │ │ API Status | Environment | Theme | User Info | Logout │   │
│ │ Toggle  │ │ Online     │ Development  │ 🌙    │ Admin    │ 🚪   │   │
│ └─────────┘ └─────────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────────────────┤
│ Sidebar (280px width) │ Main Content Area                                    │
│ ┌─────────────────────┐ │ ┌─────────────────────────────────────────────┐ │
│ │ 🏠 Dashboard        │ │ │ Dashboard Header                              │ │
│ │ 🎯 Reconnaissance   │ │ │ Security Dashboard                           │ │
│ │ ⚔️ Exploits         │ │ │ Real-time monitoring and analytics           │ │
│ │ 💣 Payloads         │ │ └─────────────────────────────────────────────┘ │
│ │ 🤖 Agents           │ │ ┌─────────────────────────────────────────────┐ │
│ │ 📊 Reports          │ │ │ Metrics Grid (4 cards)                        │ │
│ │ ⏰ Activity         │ │ │ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐         │ │
│ │ 📅 Schedule         │ │ │ │Scans│ │Explt│ │Tgts │ │Risk │         │ │
│ │ 🔔 SIEM             │ │ │ │ 142 │ │ 89  │ │ 23  │ │ 67  │         │ │
│ │ ⚙️ Admin            │ │ │ └─────┘ └─────┘ └─────┘ └─────┘         │ │
│ └─────────────────────┘ │ └─────────────────────────────────────────────┘ │
│                        │ ┌─────────────────────────────────────────────┐ │
│                        │ │ Charts Section (2x2 grid)                    │ │
│                        │ │ ┌─────────────────┐ ┌─────────────────┐ │ │
│                        │ │ │ Weekly Activity  │ │ Vulnerability    │ │
│                        │ │ │ Line Chart       │ │ Pie Chart        │ │ │
│                        │ │ └─────────────────┘ └─────────────────┘ │ │
│                        │ │ ┌─────────────────┐ ┌─────────────────┐ │ │
│                        │ │ │ Target Types     │ │ Recent Activity  │ │
│                        │ │ │ Bar Chart        │ │ Activity List    │ │ │
│                        │ │ └─────────────────┘ └─────────────────┘ │ │
│                        │ └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Login Page Design

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│                        ┌─────────────────────────────┐                       │
│                        │      Hades Security Platform    │                       │
│                        │                             │                       │
│                        │         🛡️ LOGO               │                       │
│                        │                             │                       │
│                        │    Enterprise Authentication   │                       │
│                        │                             │                       │
│                        │ ┌─────────────────────────────┐ │                       │
│                        │ │ Username                    │ │                       │
│                        │ └─────────────────────────────┘ │                       │
│                        │                             │                       │
│                        │ ┌─────────────────────────────┐ │                       │
│                        │ │ Password                    │ │                       │
│                        │ └─────────────────────────────┘ │                       │
│                        │                             │                       │
│                        │         [ LOGIN BUTTON ]        │                       │
│                        │                             │                       │
│                        └─────────────────────────────┘                       │
│                                                                                 │
│  [ Dev Access Button ] - Bottom left corner with glass-morphism effect         │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Dev Access Component

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 🔧 Dev Access                                                    │   │
│  │ Development Environment                                            │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 👨‍💻 Development Lead                                         │ │   │
│  │ │ Senior developer with full system access                      │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🧪 QA Engineer                                               │ │   │
│  │ │ Quality assurance specialist for testing frameworks           │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🔒 Security Tester                                          │ │   │
│  │ │ Security testing professional for vulnerability assessment  │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🎨 Frontend Developer                                        │ │   │
│  │ │ UI/UX developer focused on interface design                  │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ ⚙️ Backend Developer                                         │ │   │
│  │ │ Server-side developer focused on API development             │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Environment Switcher

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 🔧 Development ▼                                                │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ Switch Environment                                          │   │
│  │ │ Current: Development                                        │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🔧 Development     ✓                                        │ │   │
│  │ │ development                                               │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🧪 Testing                                                  │ │   │
│  │ │ testing                                                   │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ ✅ QA                                                       │ │   │
│  │ │ qa                                                        │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🚀 Staging                                                  │ │   │
│  │ │ staging                                                   │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🏭 Production                                               │ │   │
│  │ │ production                                                │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  │                                                                 │   │
│  │ ┌─────────────────────────────────────────────────────────────┐ │   │
│  │ │ 🔧 Development environment - safe for testing               │ │   │
│  │ └─────────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Key Visual Elements

### Glass-Morphism Effects
- Semi-transparent backgrounds with backdrop blur
- Subtle borders with rgba colors
- Smooth shadows and gradients
- Hover states with transform effects

### Typography
- **Headers**: Inter or similar sans-serif, 600-700 weight
- **Body**: Inter or similar, 400-500 weight
- **Monospace**: JetBrains Mono for code/technical data

### Icons
- SVG icons with consistent stroke width (2px)
- Professional security-themed iconography
- Consistent sizing (16px, 20px, 24px, 32px)

### Animations
- Smooth transitions (0.2s-0.3s cubic-bezier)
- Hover effects with transform and shadow changes
- Loading spinners with rotation animations
- Fade-in effects for content loading

### Responsive Design
- Mobile: Sidebar collapses to icon-only
- Tablet: Adjusted spacing and font sizes
- Desktop: Full layout with optimal spacing

## Environment-Specific Styling

### Development (Blue Theme)
- Primary accent: #3b82f6
- Dev Access: Fully visible with 5 roles
- Border colors: Blue-tinted

### Testing (Purple Theme)
- Primary accent: #8b5cf6
- Dev Access: Visible with testing roles
- Border colors: Purple-tinted

### QA (Orange Theme)
- Primary accent: #f59e0b
- Dev Access: Visible with QA roles
- Border colors: Orange-tinted

### Staging (Green Theme)
- Primary accent: #10b981
- Dev Access: Visible with staging roles
- Border colors: Green-tinted

### Production (Red Theme)
- Primary accent: #ef4444
- Dev Access: Hidden (Super Admin only)
- Border colors: Red-tinted
- Additional security warnings

This design maintains the authentic hades-toolkit aesthetic while providing a professional, modern security platform interface with comprehensive environment switching capabilities.
