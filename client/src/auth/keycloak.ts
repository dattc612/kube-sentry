import Keycloak from "keycloak-js";

const keycloak = new Keycloak({
  url: process.env.KEYCLOAK_URL || "",
  realm: process.env.KEYCLOAK_REALM || "",
  clientId: process.env.KEYCLOAK_CLIENT_ID || "",
});

export let isKeycloakInitStarted = false;

export const markKeycloakInitStarted = () => {
  isKeycloakInitStarted = true;
};

export default keycloak;
