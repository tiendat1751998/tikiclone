export function Footer() {
  return (
    <footer className="bg-white border-t border-gray-200 mt-8">
      <div className="container mx-auto px-4 py-8 text-center text-sm text-gray-500">
        &copy; {new Date().getFullYear()} Tiki Clone. All rights reserved.
      </div>
    </footer>
  );
}
