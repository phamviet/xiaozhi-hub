# Xiaozhi Hub - UI Documentation

## Overview
The UI for Xiaozhi Hub is a modern Single Page Application (SPA) built with React, Vite, and Tailwind CSS. It serves as the frontend for managing AI devices, agents, and system configurations.

## Tech Stack
- **Framework**: React 19
- **Build Tool**: Vite 7
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4 (with `tailwindcss-animate`)
- **State Management**: Nanostores
- **Routing**: @nanostores/router
- **Backend Integration**: PocketBase JS SDK
- **UI Components**: Shadcn UI (Radix UI primitives + Tailwind)
- **Icons**: Lucide React
- **Form Validation**: Valibot
- **Toast Notifications**: Sonner

## Project Structure
- `src/components/ui`: Reusable UI components (buttons, dialogs, inputs, etc.), largely based on Shadcn UI.
- `src/components/agents`: Agent-related components (`AgentList`, `AgentCard`).
- `src/components/devices`: Device management components (`BindDeviceDialog`).
- `src/components/chat`: Chat history components (`ChatHistoryDialog`, `AudioPlayer`).
- `src/components`: Application-specific components (login forms, router, theme provider).
- `src/routes`: Page components (Home, Login).
- `src/lib`: Utilities, API client, and global stores.
- `src/index.css`: Global styles, Tailwind configuration, and CSS variables for theming.

## Design System & Styling

### Theming
The application supports Light, Dark, and System themes.
- **Implementation**: CSS variables (HSL values) defined in `src/index.css`.
- **Theme Provider**: `src/components/theme-provider.tsx` manages the theme state and applies the `.dark` class to the document root.
- **Colors**:
  - `background`, `foreground`
  - `primary`, `secondary`, `accent`, `muted`
  - `destructive`
  - `card`, `popover`
  - `border`, `input`, `ring`
  - Chart colors (`--chart-1` to `--chart-5`)

### Typography
- **Font**: Inter (via `InterVariable.woff2`).
- **Base Size**: 16px (implied by Tailwind defaults).

### Layout
- **Container**: Max width `1500px` (variable `--container`).
- **Responsive Breakpoints**:
  - `xs`: 26.6rem
  - `450`: 28rem
  - `2xl`: 90rem
  - Standard Tailwind breakpoints (sm, md, lg, xl).

### Components
Components follow the Shadcn UI pattern:
- **Radix UI**: Headless, accessible primitives.
- **Tailwind**: Styled via `className` props and `cva` (class-variance-authority) for variants.
- **Utils**: `cn()` helper (clsx + tailwind-merge) is used extensively to merge classes.

## Features

### Agent Management
- **Agent List**: Displays all agents linked to the user.
- **Agent Card**: Shows basic agent info (name, language, system prompt).
- **Actions**:
  - **Configure**: (Placeholder) Modal for model configuration.
  - **Devices**: View linked devices and bind new ones.
  - **History**: View chat history.

### Device Management
- **Device List**: Shows devices linked to a specific agent.
- **Device Binding**: Allows users to bind a device using a 6-digit code via `BindDeviceDialog`.

### Chat History
- **Split View**: Conversations list on the left, chat messages on the right.
- **Audio Playback**: Support for playing audio messages directly in the chat view.
- **Grouping**: Messages are grouped by conversation ID.

## Key Libraries & Patterns

### State Management (Nanostores)
- **Atoms**: Used for global state like authentication status (`$authenticated` in `src/lib/stores.ts`).
- **Router**: `$router` (from `@nanostores/router`) manages URL state.

### Authentication
- **PocketBase**: Handles auth logic (email/password, OAuth2).
- **Flow**:
  - `src/routes/login.tsx`: Main login page.
  - `src/components/login/auth-form.tsx`: Handles login/registration forms.
  - `src/lib/api.ts`: Initializes the PocketBase client (`pb`) and provides auth helpers.

### Routing
- **Definition**: Routes are defined in `src/components/router.tsx`.
- **Navigation**: `navigate()` function or `<Link>` component.
- **Lazy Loading**: Route components are lazy-loaded in `src/main.tsx`.

## Development Guidelines

### Adding New Components
1. Prefer using existing UI components from `src/components/ui`.
2. If a new primitive is needed, consider adding it from Shadcn UI.
3. Use `cn()` for class merging.
4. Ensure dark mode compatibility by using CSS variables (e.g., `bg-background`, `text-foreground`).

### API Integration
1. Import `pb` from `@/lib/api`.
2. Use PocketBase SDK methods (e.g., `pb.collection('...').getList()`).
3. Handle errors gracefully (use `sonner` for toasts).

### Code Style
- **Formatter**: Biome (`biome.json`).
- **Linting**: Biome.
- **Imports**: Use `@/` alias for `src/` directory.

## Future UI Tasks
- Expand the dashboard to show device statistics.
- Create forms for editing AI Agents and Models.
