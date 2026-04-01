/** Shared MUI Outlined Select styling (Talents filters + student registration). */
export const aaltoOutlinedSelectSx = {
  borderRadius: "12px",
  "& .MuiOutlinedInput-root": {
    borderRadius: "12px",
    backgroundColor: "#f8fafc",
    transition: "box-shadow 0.2s ease, border-color 0.2s ease",
    "& fieldset": {
      borderColor: "#e2e8f0",
    },
    "&:hover fieldset": {
      borderColor: "#cbd5e1",
    },
    "&.Mui-focused fieldset": {
      borderColor: "#3194ee",
      borderWidth: "2px",
    },
  },
  "& .MuiSelect-select": {
    py: 1.125,
    fontSize: "0.875rem",
    fontWeight: 500,
    color: "#1e293b",
  },
};

export const aaltoSelectMenuProps = {
  PaperProps: {
    sx: {
      maxHeight: 360,
      borderRadius: "12px",
      mt: 0.75,
      boxShadow: "0 12px 40px rgba(15, 23, 42, 0.12)",
      border: "1px solid #e2e8f0",
      "& .MuiMenuItem-root": {
        fontSize: "0.8125rem",
        py: 1,
        whiteSpace: "normal",
        alignItems: "flex-start",
      },
      "& .MuiListSubheader-root": {
        fontSize: "0.7rem",
        fontWeight: 700,
        lineHeight: 1.35,
        py: 1,
        px: 2,
        bgcolor: "#f1f5f9",
        color: "#475569",
        whiteSpace: "normal",
      },
    },
  },
};
