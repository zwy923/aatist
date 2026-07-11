const UTM_KEYS = ["utm_source", "utm_medium", "utm_campaign", "utm_content", "utm_term"];

export function getTrackingFields() {
    const params = new URLSearchParams(window.location.search);
    const utm = {};
    UTM_KEYS.forEach((key) => {
        const value = params.get(key);
        if (value) utm[key] = value;
    });

    return {
        ...utm,
        referrer: document.referrer || undefined,
        user_agent: navigator.userAgent,
    };
}
