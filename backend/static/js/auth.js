/**
 * FeaturePlus Authentication Utilities
 * Handles JWT token storage, retrieval, and automatic inclusion in HTMX requests
 */

// Constants
const JWT_STORAGE_KEY = 'jwt';
const ROLE_STORAGE_KEY = 'role';
const TOKEN_EXPIRY_KEY = 'jwt_expiry';

/**
 * Authentication utility functions
 */
const Auth = {
  /**
   * Store JWT token in localStorage with expiry information
   * @param {string} token - JWT token to store
   * @param {number} expiryInSeconds - Optional token expiry in seconds (default: 72 hours)
   */
  storeToken: function(token, expiryInSeconds = 72 * 60 * 60) {
    if (token) {
      localStorage.setItem(JWT_STORAGE_KEY, token);
      
      // Calculate and store expiry timestamp
      const expiryTime = Date.now() + (expiryInSeconds * 1000);
      localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
      
      console.log('JWT token stored with expiry:', new Date(expiryTime).toLocaleString());
    }
  },

  /**
   * Store user role in localStorage
   * @param {string} role - User role (manager/developer)
   */
  storeRole: function(role) {
    if (role) {
      localStorage.setItem(ROLE_STORAGE_KEY, role);
    }
  },

  /**
   * Get stored JWT token, checking for expiry
   * @returns {string|null} JWT token or null if not found or expired
   */
  getToken: function() {
    const token = localStorage.getItem(JWT_STORAGE_KEY);
    if (!token) return null;
    
    // Check if token is expired
    if (this.isTokenExpired()) {
      console.warn('Token has expired, clearing authentication data');
      this.logout();
      return null;
    }
    
    return token;
  },

  /**
   * Check if the stored token is expired
   * @returns {boolean} True if token is expired or expiry info is missing
   */
  isTokenExpired: function() {
    const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);
    if (!expiryTime) return true; // No expiry info, consider expired
    
    return Date.now() > parseInt(expiryTime, 10);
  },

  /**
   * Get stored user role
   * @returns {string|null} User role or null if not found
   */
  getRole: function() {
    return localStorage.getItem(ROLE_STORAGE_KEY);
  },

  /**
   * Check if user is authenticated (has valid JWT token)
   * @returns {boolean} True if authenticated with non-expired token
   */
  isAuthenticated: function() {
    return !!this.getToken(); // getToken already checks for expiry
  },

  /**
   * Clear authentication data (logout)
   * @param {boolean} redirect - Whether to redirect to login page (default: true)
   */
  logout: function(redirect = true) {
    localStorage.removeItem(JWT_STORAGE_KEY);
    localStorage.removeItem(ROLE_STORAGE_KEY);
    localStorage.removeItem(TOKEN_EXPIRY_KEY);
    
    // Redirect to login page if requested
    if (redirect) {
      window.location.href = '/web/login';
    }
  },
  
  /**
   * Parse JWT token to extract payload
   * @param {string} token - JWT token to parse
   * @returns {Object|null} Decoded payload or null if invalid
   */
  parseToken: function(token) {
    try {
      if (!token) return null;
      
      // Get the payload part (second segment) of the JWT
      const base64Url = token.split('.')[1];
      if (!base64Url) return null;
      
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      }).join(''));
      
      return JSON.parse(jsonPayload);
    } catch (e) {
      console.error('Error parsing JWT token:', e);
      return null;
    }
  }
};

/**
 * Initialize HTMX authentication hooks
 */
document.addEventListener('DOMContentLoaded', function() {
  // Add Authorization header to all HTMX requests to /web/ and /api/ routes
  htmx.on('htmx:configRequest', function(event) {
    const token = Auth.getToken();
    
    // Check if request URL starts with /web/ or /api/
    const url = event.detail.path || event.detail.url;
    if (token && (url.startsWith('/web/') || url.startsWith('/api/'))) {
      event.detail.headers['Authorization'] = `Bearer ${token}`;
      // Add credentials mode to ensure cookies are sent
      event.detail.xhr.withCredentials = true;
    }
  });

  // Handle response errors including 401 Unauthorized
  htmx.on('htmx:responseError', function(event) {
    const xhr = event.detail.xhr;
    
    if (xhr.status === 401) {
      console.warn('Received 401 Unauthorized response');
      
      // Check response for more specific error information
      try {
        const response = JSON.parse(xhr.responseText);
        if (response && response.code) {
          console.warn('Auth error code:', response.code);
          
          // Handle specific error codes
          if (response.code === 'token_expired') {
            console.warn('Token expired, redirecting to login');
          }
        }
      } catch (e) {
        // Response wasn't JSON or couldn't be parsed
      }
      
      // Clear auth data and redirect to login
      Auth.logout();
    } else if (xhr.status === 403) {
      console.warn('Received 403 Forbidden response - insufficient permissions');
      // Could show a permission denied message instead of logout
    }
  });
  
  // Check token validity on page load
  if (Auth.isTokenExpired() && Auth.getToken()) {
    console.warn('Token expired on page load, logging out');
    Auth.logout();
  }
});

// Export Auth object for use in other scripts
window.Auth = Auth;
