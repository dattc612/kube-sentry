export interface LoginResponse {
  token: string;
}

export interface KeycloakConfig {
  url: string;
  realm: string;
  clientId: string;
}
export interface AppConfig {
  auth: {
    keycloak: KeycloakConfig;
  };
}
