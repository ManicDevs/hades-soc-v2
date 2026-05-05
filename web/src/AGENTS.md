# 🎨 Frontend Dashboard - Security & Typing Rules

## 1. 🛡️ API Client Security
- **Token Management:** Never store JWT tokens in localStorage. Use secure httpOnly cookies or memory storage with auto-refresh.
- **CSP Compliance:** All inline scripts must have proper nonce attributes. No dynamic script injection.
- **API Validation:** Always validate API responses against the TypeScript schema before rendering.

## 2. 📝 Strict Typing Requirements
- **No `any` Types:** All API responses must be typed. Use discriminative unions for SIEM event types.
- **Schema Alignment:** Before updating Go API handlers, verify types match `web/dashboard/src/types/schema.ts`.
- **Error Boundaries:** Every async component must be wrapped in error boundaries with proper fallback UI.

## 3. 🔄 JSON Schema Alignment
- **Type Generation:** Run `npm run generate-types` after any Go API schema changes.
- **Validation:** Use `zod` or similar for runtime validation of API responses.
- **Mock Data:** All mock data for development must match the exact JSON schema from the backend.

## 4. 🔐 Security Best Practices
- **XSS Prevention:** All user input must be sanitized using `DOMPurify` before rendering.
- **CSRF Protection:** Include CSRF tokens in all state-changing API requests.
- **Secure Headers:** Ensure all API requests include proper security headers (X-Content-Type-Options, etc.).

## 5. 📊 Dashboard Performance
- **Virtual Scrolling:** For large datasets (logs, events), implement virtual scrolling to prevent memory leaks.
- **Lazy Loading:** Chart components and heavy visualizations must be lazy-loaded.
- **WebSocket Security:** WebSocket connections must use WSS and implement proper reconnection logic with exponential backoff.

## 6. ✅ Frontend Definition of Done
1. **Type Safety:** All API responses are properly typed and validated
2. **Security Review:** No XSS vulnerabilities, proper CSP implementation
3. **Performance:** Bundle size optimized, lazy loading implemented
4. **Schema Sync:** Frontend types match backend Go struct definitions
