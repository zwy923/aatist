export const aggregateTags = (opportunities, limit = 6) => {
  const counts = opportunities.reduce((acc, opp) => {
    (opp.tags || []).forEach((tag) => {
      const normalized = tag.trim();
      if (!normalized) return;
      acc[normalized] = (acc[normalized] || 0) + 1;
    });
    return acc;
  }, {});

  return Object.entries(counts)
    .map(([tag, count]) => ({ tag, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, limit);
};

