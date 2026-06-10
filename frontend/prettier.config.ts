/**
 * @see https://prettier.io/docs/configuration
 * @type {import("prettier").Config}
 */
const config = {
  trailingComma: 'es5',
  tabWidth: 2,
  singleQuote: true,
  endOfLine: 'auto',
  printWidth: 80,
  useTabs: false,
  semi: true,
  bracketSpacing: true,
  bracketSameLine: false,
  arrowParens: 'avoid',
  importOrder: [
    'react',
    '<THIRD_PARTY_MODULES>',
    '^@/components/(.*)$',
    '^@/contexts/(.*)$',
    '^@/hooks/(.*)$',
    '^@/stores/(.*)$',
    '^@/styles/(.*)$',
    '^@/schemas/(.*)$',
    '^@/lib/(.*)$',
    '^@/utils/(.*)$',
    '^[./]',
  ],
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
  plugins: [
    '@trivago/prettier-plugin-sort-imports',
    'prettier-plugin-tailwindcss',
  ],
};

export default config;
