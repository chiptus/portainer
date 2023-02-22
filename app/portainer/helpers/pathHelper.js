/**
 * calculates baseHref
 *
 * return {string} [base href]
 *
 */
export function baseHref() {
  const base = document.getElementById('base');
  return base ? base.getAttribute('href') || '/' : '/';
}
