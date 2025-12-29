export default function CompanyLoading() {
  return (
    <div className="container mx-auto p-6">
      <div className="space-y-6">
        {/* Skeleton для заголовка */}
        <div className="bg-white rounded-lg border p-6">
          <div className="animate-pulse">
            <div className="flex items-start justify-between mb-4">
              <div className="flex-1">
                <div className="h-8 bg-muted rounded w-3/4 mb-2"></div>
                <div className="h-6 bg-muted rounded w-1/2 mb-4"></div>
              </div>
              <div className="flex gap-2">
                <div className="h-8 w-24 bg-muted rounded"></div>
                <div className="h-8 w-32 bg-muted rounded"></div>
                <div className="h-8 w-28 bg-muted rounded"></div>
              </div>
            </div>
            <div className="grid grid-cols-3 gap-6">
              <div className="space-y-2">
                <div className="h-4 bg-muted rounded w-16"></div>
                <div className="h-6 bg-muted rounded w-32"></div>
              </div>
              <div className="space-y-2">
                <div className="h-4 bg-muted rounded w-12"></div>
                <div className="h-6 bg-muted rounded w-28"></div>
              </div>
              <div className="space-y-2">
                <div className="h-4 bg-muted rounded w-14"></div>
                <div className="h-6 bg-muted rounded w-24"></div>
              </div>
            </div>
          </div>
        </div>
        
        {/* Skeleton для табов */}
        <div className="bg-white rounded-lg border">
          <div className="animate-pulse">
            <div className="flex border-b">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="flex-1 p-4">
                  <div className="h-4 bg-muted rounded"></div>
                </div>
              ))}
            </div>
            <div className="p-6 space-y-4">
              <div className="grid gap-6 md:grid-cols-2">
                <div className="h-48 bg-muted rounded"></div>
                <div className="h-48 bg-muted rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}