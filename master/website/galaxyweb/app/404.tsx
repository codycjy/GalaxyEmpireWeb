import { Button } from "@/components/ui/button";
import Link from "next/link";

const Custom404 = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background">
      <div className="text-center space-y-4">
        <h1 className="text-6xl font-bold text-primary">404</h1>
        <h2 className="text-2xl font-semibold text-foreground">Page Not Found</h2>
        <p className="text-muted-foreground">
          Sorry, the page you are looking for does not exist.
        </p>
        <Link href="/">
          <Button>
            Return to Home
          </Button>
        </Link>
      </div>
    </div>
  );
};

export default Custom404;