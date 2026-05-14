import { Route, Router, useNavigate } from "@solidjs/router";
import { JSX, createEffect, onMount } from "solid-js";
import { authStore } from "./stores/auth";
import { SplashScreen } from "./pages/Splash";
import { LoginPage } from "./pages/login";
import { RegisterPage } from "./pages/Register";
import { ChatPage } from "./pages/chat";
import { SettingsPage } from "./pages/settings";

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
            <ChatPage />
          </ProtectedRoute>
        )}
      />
      <Route
        path="/settings"
        component={() => (
          <ProtectedRoute>
            <SettingsPage />
          </ProtectedRoute>
        )}
      />
    </Router>
  );
}