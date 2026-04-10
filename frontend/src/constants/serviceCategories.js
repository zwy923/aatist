/**
 * Broad / specific service options (keep in sync with ServicesSection form).
 */
export const BROAD_CATEGORIES = {
  Design: ["Logo Design", "Brand Identity", "Illustration", "Print Design", "Presentation Design", "Infographics"],
  "Graphics & Design": ["Logo & Brand Identity", "Illustration & Drawing", "Print Design", "Presentation Design"],
  "Website & Digital": ["Web Design", "UI/UX Design", "Digital Product"],
  "Video & Animation": ["Video Editing", "Animation", "Motion Graphics", "Explainer Videos"],
  Photography: ["Event Photography", "Portrait Photography", "Product Photography", "Food Photography"],
};

export const SPECIFIC_SERVICE_OTHER = "Other";

export function isPresetSpecific(broad, spec) {
  const presets = BROAD_CATEGORIES[broad] || [];
  return !!(spec && presets.includes(spec));
}

/**
 * Profile / cards: preset → show specific only; custom ("Other") → show saved service title, else fall back.
 */
export function getProfileServiceHeading(service) {
  const title = (service?.title || "").trim();
  const cat = (service?.category || "").trim();
  if (!cat && !title) return "Service";
  const idx = cat.indexOf(" > ");
  if (idx === -1) return title || cat || "Service";
  const broad = cat.slice(0, idx).trim();
  const spec = cat.slice(idx + 3).trim();
  if (isPresetSpecific(broad, spec)) return spec;
  return title || spec || cat || "Service";
}
