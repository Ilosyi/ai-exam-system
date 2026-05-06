/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          navy: "#041833",
          blue: "#1f4cd4"
        }
      }
    }
  },
  plugins: []
};
