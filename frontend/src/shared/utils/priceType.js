/**
 * price_type is stored as comma-separated tokens: hourly, project, negotiable (fixed order when encoding).
 */
export function parsePriceTypeTokens(priceType) {
  if (priceType == null || priceType === "") return [];
  return [
    ...new Set(
      String(priceType)
        .split(",")
        .map((t) => t.trim().toLowerCase())
        .filter(Boolean)
    ),
  ];
}

export function encodePriceTypePayload({ hourly, project, negotiable }) {
  const parts = [];
  if (hourly) parts.push("hourly");
  if (project) parts.push("project");
  if (negotiable) parts.push("negotiable");
  if (parts.length === 0) parts.push("negotiable");
  return parts.join(",");
}

/** One-line summary for cards and profile (e.g. "€50–80 / hr · Negotiable"). */
export function formatServicePriceLine(service) {
  const tokens = parsePriceTypeTokens(service?.price_type);
  if (!tokens.length) tokens.push("negotiable");
  const bits = [];
  const pmin = service?.price_min;
  const pmax = service?.price_max;
  const euroRange = () => {
    if (pmin != null && pmax != null) return `€${pmin}–€${pmax}`;
    if (pmin != null) return `€${pmin}`;
    if (pmax != null) return `€${pmax}`;
    return null;
  };
  if (tokens.includes("hourly")) {
    const r = euroRange();
    bits.push(r ? `${r} / hr` : "Hourly");
  }
  if (tokens.includes("project")) {
    const r = euroRange();
    bits.push(r ? `${r} / project` : "Project");
  }
  if (tokens.includes("negotiable")) bits.push("Negotiable");
  return bits.join(" · ");
}
