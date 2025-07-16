/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
      "./backend/templates/**/*.html",
    ],
    theme: {
      extend: {
        colors: {
          primary: {
            DEFAULT: '#4f46e5', // Corresponds to indigo-600
            '600': '#4f46e5',   // Explicitly define for consistency
            '700': '#4338ca'    // Darker shade for hover, corresponds to indigo-700
          },
          success: {
            DEFAULT: '#10b981', // Corresponds to green-500
            '100': '#d1fae5'   // Light background for alerts/badges
          },
          warning: {
            DEFAULT: '#f59e0b', // Corresponds to amber-500
            '100': '#fef3c7'   // Light background
          },
          danger: {
            DEFAULT: '#ef4444', // Corresponds to red-500
            '100': '#fee2e2'   // Light background
          }
        }
      }
    },
    plugins: [],
  }