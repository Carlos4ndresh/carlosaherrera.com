const tailwind = require('../tailwind')

module.exports = {
  pathPrefix: '/', // Prefix for all links. If you deploy your site to example.com/portfolio your pathPrefix should be "/portfolio"

  siteTitle: 'Carlos A. Herrera - Developer at night, Solutions Architect ther rest of the time', // Navigation and Site Title
  siteTitleAlt: 'carlos4ndresh', // Alternative Site title for SEO
  siteTitleShort: 'Carlosandresh', // short_name for manifest
  siteHeadline: 'Creating marvelous art & blazginly fast websites', // Headline for schema.org JSONLD
  siteUrl: 'https://www.carlosaherrera.com', // Domain of your site. No trailing slash!
  siteLanguage: 'en', // Language Tag on <html> element
  siteLogo: '/src/images/monphoto_tn.jpg', // Used for SEO and manifest
  siteDescription: '',
  author: 'Carlos A. Herrera', // Author for schema.org JSONLD

  userTwitter: '@Carlos4ndresh', // Twitter Username
  googleAnalyticsID: 'UA-138228571-1',

  // Manifest and Progress color
  themeColor: tailwind.colors.orange,
  backgroundColor: tailwind.colors.blue,
}
