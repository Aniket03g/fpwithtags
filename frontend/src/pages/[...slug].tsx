import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';

export default function DynamicPage() {
  const router = useRouter();
  const { slug } = router.query;

  // Debug: log when the catch-all route loads and what the slug is
  console.log('[CatchAll] Loaded [...slug].tsx, slug:', slug);

  // Build the path for the JSON file
  let slugPath = '';
  if (Array.isArray(slug)) {
    slugPath = slug.join('/');
  } else if (typeof slug === 'string') {
    slugPath = slug;
  }
  const jsonUrl = `/data/${slugPath}.json`;

  // Debug: log the JSON URL being fetched
  console.log('[CatchAll] JSON URL to fetch:', jsonUrl);

  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    if (!slugPath) {
      console.log('[CatchAll] No slugPath, skipping fetch');
      return;
    }
    setLoading(true);
    setNotFound(false);
    setData(null);
    console.log('[CatchAll] Fetching:', jsonUrl);
    fetch(jsonUrl)
      .then((res) => {
        if (!res.ok) throw new Error('Not Found');
        return res.json();
      })
      .then((json) => {
        console.log('[CatchAll] Fetched JSON:', json);
        setData(json);
        setLoading(false);
      })
      .catch((err) => {
        console.log('[CatchAll] Fetch error:', err);
        setNotFound(true);
        setLoading(false);
      });
  }, [jsonUrl, slugPath]);

  if (loading) return <div>Loading...</div>;
  if (notFound) return <div>Not Found</div>;

  // Helper to render objects
  const renderObject = (obj: any) => (
    <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
      <h1>{obj.title || obj.name || 'Untitled'}</h1>
      <div>{obj.description || obj.content || JSON.stringify(obj, null, 2)}</div>
    </div>
  );

  // If data is an array, render a list
  if (Array.isArray(data)) {
    return (
      <div>
        <h1>List</h1>
        <ul>
          {data.map((item, idx) => (
            <li key={item.id || idx}>{renderObject(item)}</li>
          ))}
        </ul>
      </div>
    );
  }

  // If data is an object, render it
  if (typeof data === 'object' && data !== null) {
    return renderObject(data);
  }

  // Fallback
  return <div>Invalid data</div>;
}

// For static export compatibility
export async function getStaticPaths() {
  return {
    paths: [], // No pre-generated paths
    fallback: false, // All dynamic routes handled client-side
  };
}

export async function getStaticProps() {
  return { props: {} };
} 