export function getOpportunityCategories(opportunity) {
  const primary = typeof opportunity?.category === "string" ? opportunity.category : "";
  const tags = Array.isArray(opportunity?.tags) ? opportunity.tags : [];
  const merged = [primary, ...tags];
  const seen = new Set();
  const categories = [];

  for (const item of merged) {
    const value = typeof item === "string" ? item.trim() : "";
    if (!value) continue;
    const key = value.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    categories.push(value);
  }

  return categories;
}
