const WEIGHTS = [5, 4, 3, 2, 7, 6, 5, 4, 3, 2] as const;
const VALID_PREFIXES = new Set(['10', '15', '17', '20']);

/** Valida un RUC peruano con el dígito verificador de SUNAT (misma regla que el backend). */
export function isValidRuc(value: string): boolean {
  if (!/^\d{11}$/.test(value) || !VALID_PREFIXES.has(value.slice(0, 2))) {
    return false;
  }
  const sum = WEIGHTS.reduce((acc, weight, i) => acc + Number(value[i]) * weight, 0);
  let digit = 11 - (sum % 11);
  if (digit === 10) digit = 0;
  if (digit === 11) digit = 1;
  return digit === Number(value[10]);
}
