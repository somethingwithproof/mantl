import React, { useState, useEffect } from 'react';
import './App.css';

interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  category: string;
  stock: number;
  sku: string;
  image_url?: string;
}

const API_BASE_URL = process.env.REACT_APP_API_URL || '/api';

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [categories, setCategories] = useState<string[]>([]);

  // Fetch categories
  useEffect(() => {
    fetch(`${API_BASE_URL}/products/api/v1/categories`)
      .then(res => res.json())
      .then(data => setCategories(['all', ...data]))
      .catch(err => console.error('Failed to load categories:', err));
  }, []);

  // Fetch products
  useEffect(() => {
    setLoading(true);
    setError(null);

    const url = selectedCategory === 'all'
      ? `${API_BASE_URL}/products/api/v1/products`
      : `${API_BASE_URL}/products/api/v1/products?category=${selectedCategory}`;

    fetch(url)
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch products');
        return res.json();
      })
      .then(data => {
        setProducts(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [selectedCategory]);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(price);
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>🛒 E-Commerce Store</h1>
        <p className="subtitle">Powered by mantl Platform</p>
      </header>

      <nav className="category-nav">
        <h3>Categories</h3>
        <div className="category-buttons">
          {categories.map(cat => (
            <button
              key={cat}
              className={selectedCategory === cat ? 'active' : ''}
              onClick={() => setSelectedCategory(cat)}
            >
              {cat.charAt(0).toUpperCase() + cat.slice(1)}
            </button>
          ))}
        </div>
      </nav>

      <main className="product-grid">
        {loading && (
          <div className="loading">
            <div className="spinner"></div>
            <p>Loading products...</p>
          </div>
        )}

        {error && (
          <div className="error">
            <h3>⚠️ Error</h3>
            <p>{error}</p>
          </div>
        )}

        {!loading && !error && products.length === 0 && (
          <div className="empty">
            <h3>No Products Found</h3>
            <p>Try selecting a different category</p>
          </div>
        )}

        {!loading && !error && products.map(product => (
          <div key={product.id} className="product-card">
            {product.image_url ? (
              <img src={product.image_url} alt={product.name} />
            ) : (
              <div className="placeholder-image">📦</div>
            )}
            <div className="product-info">
              <h3>{product.name}</h3>
              <p className="category">{product.category}</p>
              <p className="description">{product.description}</p>
              <div className="product-footer">
                <span className="price">{formatPrice(product.price)}</span>
                <span className={`stock ${product.stock > 0 ? 'in-stock' : 'out-of-stock'}`}>
                  {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
                </span>
              </div>
              <button
                className="add-to-cart"
                disabled={product.stock === 0}
              >
                {product.stock > 0 ? 'Add to Cart' : 'Sold Out'}
              </button>
            </div>
          </div>
        ))}
      </main>

      <footer className="App-footer">
        <p>mantl E-Commerce Platform Demo</p>
        <p className="tech-stack">
          React · FastAPI · PostgreSQL · Kubernetes
        </p>
      </footer>
    </div>
  );
}

export default App;
