import { Route, Router, useNavigate } from "@solidjs/router";
import { JSX, createEffect } from "solid-js";
import { authStore } from "./stores/auth";
import { appNav } from "./services/nav";
import { SplashScreen } from "./pages/Splash";
import { LoginPage } from "./pages/login";
import { RegisterPage } from "./pages/Register";
import { ChatPage } from "./pages/chat";
import { SettingsPage } from "./pages/settings";
import { DownloadProgress } from "./components/ui/download-progress";

function ProtectedRoute(props: { children: JSX.Element }) {
  const navigate = useNavigate();

  createEffect(() => {
    if (!authStore.isLoggedIn()) {
      navigate("/login", { replace: true });
    }
  });

  if (!authStore.isLoggedIn()) {
    return null;
  }

  return <>{props.children}</>;
}

/** Main app shell: keeps chat page always mounted, overlays settings with display */
function AppShell() {
  return (
    <>
      <div style={{ display: appNav.page() === "chat" ? "" : "none" }}>
        <ChatPage />
      </div>
      <div style={{ display: appNav.page() === "settings" ? "" : "none" }}>
        <SettingsPage />
      </div>
      <DownloadProgress />
    </>
  );
}

export default function App() {
  return (
    <Router>
      <Route path="/" component={SplashScreen} />
      <Route path="/login" component={LoginPage} />
      <Route path="/register" component={RegisterPage} />
      <Route
        path="/chat"
        component={() => (
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        )}
      />
      <Route
        path="/settings"
        component={() => (
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        )}
      />
    </Router>
  );
}