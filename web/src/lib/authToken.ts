let token: string | null = null

export function setAuthToken(nextToken: string | null) {
  token = nextToken || null
}

export function getAuthToken() {
  return token
}

export function clearAuthToken() {
  token = null
}
