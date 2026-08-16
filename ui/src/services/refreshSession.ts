import axios from "axios";

// The one refresh in flight, shared by every caller that needs it.
let inFlight: Promise<string> | null = null;

const clearSession = () => {
  localStorage.removeItem("accessToken");
  localStorage.removeItem("refreshToken");
  localStorage.removeItem("user");
};

/**
 * Refreshes the session, collapsing concurrent callers onto a single request.
 *
 * The server rotates refresh tokens: using one revokes it. So when several
 * requests get a 401 at the same time — which is exactly what a bulk operation
 * produces — each refreshing independently means the first succeeds and every
 * other one presents a token that no longer exists, gets rejected, and logs the
 * user out mid-operation. Sharing one promise means they all wait for the same
 * result, and only a genuine failure ends the session.
 */
export const refreshSession = (): Promise<string> => {
  if (inFlight) return inFlight;

  const refreshToken = localStorage.getItem("refreshToken");
  if (!refreshToken) {
    clearSession();
    return Promise.reject(new Error("No refresh token"));
  }

  inFlight = axios
    .post("/api/auth/refresh", { refresh_token: refreshToken })
    .then((response) => {
      const {
        access_token,
        refresh_token: newRefreshToken,
        user,
      } = response.data;

      localStorage.setItem("accessToken", access_token);
      localStorage.setItem("refreshToken", newRefreshToken);
      localStorage.setItem("user", JSON.stringify(user));

      return access_token as string;
    })
    .catch((error) => {
      clearSession();
      throw error;
    })
    .finally(() => {
      inFlight = null;
    });

  return inFlight;
};

/** Sends the user to the login screen, once, however many callers ask. */
export const redirectToLogin = () => {
  if (window.location.pathname !== "/login") {
    window.location.href = "/login";
  }
};
