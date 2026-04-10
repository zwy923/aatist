/**
 * Hire Talent filter groups + Add Service broad/specific options (single source of truth).
 * "Other" is only in the Add Service UI (ServicesSection), not in these preset lists.
 */
export const HIRE_TALENT_SERVICE_CATEGORIES = [
  {
    main: "Graphic & Illustration",
    items: [
      "Graphic & Illustration",
      "Pitch deck design",
      "Poster Design",
      "Social media post design",
      "Print & Packaging",
    ],
  },
  {
    main: "Branding",
    items: ["LOGO Design", "Brand Design"],
  },
  {
    main: "Web Design",
    items: ["Website Design", "APP Design"],
  },
  {
    main: "Photography",
    items: ["Product Photography", "Team Photography", "Event Photography"],
  },
  {
    main: "Video & Motion",
    items: ["Startup Promo Video", "Animated Video", "Video Editing", "Short Video", "Event Video"],
  },
  {
    main: "Creative Styling",
    items: ["Team Outfit Design"],
  },
  {
    main: "Exhibition & Spatial Design",
    items: ["Exhibition Design", "Booth Design"],
  },
];

/** Broad category → specific presets (no "Other" here; form adds it separately). */
export const BROAD_CATEGORIES = Object.fromEntries(
  HIRE_TALENT_SERVICE_CATEGORIES.map(({ main, items }) => [main, [...items]])
);

export const ALL_HIRE_TALENT_SERVICE_SUGGESTIONS = [
  ...new Set(HIRE_TALENT_SERVICE_CATEGORIES.flatMap((g) => g.items)),
];

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
