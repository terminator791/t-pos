# Project Structure Guide

This document provides a comprehensive overview of the project's folder structure and guidelines for developers. Understanding this structure is key to maintaining code quality and contributing effectively.

## Root Directory

The root directory contains the core application source, configuration files, and project metadata.

### Folder & File Tree

Here is a high-level overview of the main directories and files:

```
.
├── .gitignore
├── index.html
├── package.json
├── postcss.config.js
├── README.md
├── tailwind.config.js
├── vite.config.js
├── public/
│   └── favicon.svg
└── src/
    ├── App.jsx
    ├── main.jsx
    ├── assets/
    ├── components/
    ├── config/
    ├── hooks/
    ├── Layout/
    ├── mocks/
    ├── pages/
    ├── server/
    └── store/
```

### Configuration Files

The following configuration files are located at the root level:

- `vite.config.js`: Configuration for the Vite build tool.
- `tailwind.config.js`: Configuration file for the Tailwind CSS framework.
- `postcss.config.js`: Configuration for PostCSS processing.
- `package.json`: Contains project metadata, scripts, and dependencies.
- `.gitignore`: Specifies files and directories that Git should ignore.

---

## The `src` Directory

All of the main source code for the application resides within the `src` directory.

- `main.jsx`: The main entry point of the application. It renders the `App` component to the DOM.
- `App.jsx`: The root component of the application, where routing and global providers are set up.

### `src/assets`

This folder holds all static assets required by the application.

- `images/`: Contains all image assets (e.g., logos, backgrounds).
- `scss/`: Contains global SCSS files, variables, mixins, and stylesheets.

### `src/components`

This directory houses all reusable React components, organized by their function.

- `partials/`: Contains larger, partial components that make up a page layout (e.g., `Header`, `Sidebar`, `Footer`).
- `ui/`: Includes small, general-purpose UI components (e.g., `Button`, `Card`, `Dropdown`, `Input`).
- `widgets/`: Contains more complex components that often display data, like charts or statistics cards.

### `src/pages`

Contains the main view components for each page or route in the application.

- `dashboard/`: Files and components related to the dashboard page.
- `chart/`: Files and components related to the chart examples page.

### `src/store`

Contains all files related to state management using **Redux Toolkit**. This is where you will find state slices, actions, and the main store configuration.

### `src/hooks`

This directory is for custom React hooks that can be shared across multiple components (e.g., `useApi`, `useTheme`).

### `src/mocks` & `src/server`

These directories are used for development and testing purposes.

- `mocks/`: Contains mock data and constants used for populating components or pages. Examples include `data.js`, `appex-chart.js`.
- `server/`: Contains a fake API or mock server setup (e.g., using MirageJS or MSW) to simulate backend responses during development.

---

## Developer Guidelines

### Working with Components

To maintain a clean and organized codebase, please follow these conventions:

1.  **Create Reusable UI:** All general-purpose, reusable components like buttons, modals, or inputs should be placed in `src/components/ui`.
2.  **Component Structure:** Each component should be in its own folder with an `index.jsx` file. For example:
    - `src/components/ui/Button/index.jsx`
3.  **Importing Components:** Use absolute paths or aliases configured in `vite.config.js`

### Adding New Pages

To add a new page to the application:

1.  Create a new folder inside `src/pages` (e.g., `src/pages/Profile/`).
2.  Create the main page component inside this new folder (e.g., `index.jsx`).
3.  Add the new route to your router configuration file, linking the path to your newly created page component.
