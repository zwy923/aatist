import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import useAuthStore from "../shared/stores/authStore";
import { profileApi, portfolioApi } from "../features/profile/api/profile";
import PageLayout from "../shared/components/PageLayout";
import { StateContainer } from "../shared/components/ui/StateContainer";
import ProfileView from "../components/profile/ProfileView";

export default function PublicProfile() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const myId = user?.id ?? user?.user_id;
  const isOwnProfile = isAuthenticated && id && Number(id) === Number(myId);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [profile, setProfile] = useState(null);
  const [portfolio, setPortfolio] = useState([]);
  const [services, setServices] = useState([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const [profileRes, portfolioRes] = await Promise.allSettled([
          profileApi.getPublicProfile(id),
          portfolioApi.getUserPortfolio(id),
        ]);

        if (profileRes.status === "fulfilled") {
          const data = profileRes.value.data.data;
          setProfile(data);
          setServices(data?.services || []);
        } else {
          throw new Error("Failed to load user profile");
        }

        if (portfolioRes.status === "fulfilled") {
          const data = portfolioRes.value.data.data;
          setPortfolio(data?.projects || data?.items || []);
        }
      } catch (err) {
        console.error("Public profile load error:", err);
        setError("Could not load user profile");
      } finally {
        setLoading(false);
      }
    };

    if (id) fetchData();
  }, [id]);

  const goProfile = (state) => navigate("/profile", { state });

  const roleLower = profile?.role?.toLowerCase?.();
  const isOrgProfile = roleLower === "org_person" || roleLower === "org_team";
  const showTalentSections =
    !isOrgProfile &&
    (roleLower === "student" ||
      roleLower === "alumni" ||
      (services?.length ?? 0) > 0 ||
      (portfolio?.length ?? 0) > 0);

  return (
    <PageLayout noContainer>
      <StateContainer loading={loading} error={error}>
        <ProfileView
          profile={profile}
          profileUserId={id ? Number(id) : undefined}
          portfolio={portfolio}
          services={services}
          isOwnProfile={isOwnProfile}
          showServicesAndPortfolio={showTalentSections}
          bannerUrl={profile?.banner_url}
          onMessage={
            isAuthenticated && !isOwnProfile ? () => navigate(`/messages?user=${id}`) : undefined
          }
          onChangeBackground={isOwnProfile ? () => goProfile({ openBannerPicker: true }) : undefined}
          onEditProfile={isOwnProfile ? () => goProfile({ openProfileEdit: true }) : undefined}
          onAddService={isOwnProfile ? () => goProfile({ openServices: true }) : undefined}
          onAddProject={isOwnProfile ? () => goProfile({ openPortfolio: true }) : undefined}
          onPortfolioEdit={
            isOwnProfile ? (projectId) => goProfile({ editPortfolioId: projectId }) : undefined
          }
        />
      </StateContainer>
    </PageLayout>
  );
}
