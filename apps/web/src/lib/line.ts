import liff from "@line/liff";
import { setDevelopmentLineUser, setLineIdentityToken } from "../api/client";

const developmentUserID = "local-development-user";

export async function initializeLineIdentity(): Promise<boolean> {
  const liffID = import.meta.env.VITE_LIFF_ID?.trim();
  if (!liffID) {
    setDevelopmentLineUser(developmentUserID);
    return true;
  }

  await liff.init({ liffId: liffID, withLoginOnExternalBrowser: true });
  if (!liff.isLoggedIn()) {
    liff.login({ redirectUri: window.location.href });
    return false;
  }

  const idToken = liff.getIDToken();
  if (!idToken) {
    throw new Error("LINE identity token is unavailable");
  }
  setLineIdentityToken(idToken);
  return true;
}
