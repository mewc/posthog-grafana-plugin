import baseConfig from './.config/eslint.config.mjs';

export default [
  { ignores: ['dist/**', '.config/**'] },
  ...baseConfig,
];
