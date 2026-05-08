let token = null

export function setAuthToken(nextToken) {
  token = nextToken || null
}

export function getAuthToken() {
  return token
}

export function clearAuthToken() {
  token = null
}
