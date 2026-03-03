import { Hero } from '@/components/home/hero';
import { FeaturedGames } from '@/components/home/featured-games';
import { FeaturedMusic } from '@/components/home/featured-music';
import { CommunityFeed } from '@/components/home/community-feed';
import { StudioIntro } from '@/components/home/studio-intro';
import { Newsletter } from '@/components/home/newsletter';
import { Header } from '@/components/layout/header';
import { Footer } from '@/components/layout/footer';

export default function HomePage() {
  return (
    <div className="min-h-screen">
      <Header />
      <main>
        <Hero />
        <FeaturedGames />
        <FeaturedMusic />
        <CommunityFeed />
        <StudioIntro />
        <Newsletter />
      </main>
      <Footer />
    </div>
  );
}
