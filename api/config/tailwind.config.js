const defaultTheme = require("tailwindcss/defaultTheme");

module.exports = {
  content: [
    "./app/helpers/**/*.rb",
    "./app/javascript/**/*.js",
    "./app/views/**/*.{erb,haml,html,slim}",
    "./app/views/**/*.html.erb",
    "./app/views/**/*.html.haml",
    "app/views/**/*.html.haml",
    "./app/javascript/**/*.js",
    "app/javascripts/**/*.vue",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter var", ...defaultTheme.fontFamily.sans],
      },
    },
  },
  plugins: [
    require("@tailwindcss/forms"),
    require("@tailwindcss/aspect-ratio"),
    require("@tailwindcss/typography"),
  ],
};
