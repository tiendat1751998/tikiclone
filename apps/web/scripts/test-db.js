const mysql = require('mysql2/promise');
async function test() {
  const host = process.env.MYSQL_HOST || 'mysql-primary';
  const port = process.env.MYSQL_PORT || '3306';
  const user = process.env.MYSQL_USER || 'shopee';
  const pass = process.env.MYSQL_PASSWORD || 'tiki_dev';
  const db = process.env.MYSQL_DATABASE || 'tiki_platform';
  console.log(`Connecting to ${host}:${port} as ${user}...`);
  try {
    const conn = await mysql.createConnection({host, port: parseInt(port), user, password: pass, database: db});
    console.log('MySQL connected OK');
    const [r] = await conn.execute('SELECT COUNT(*) as cnt FROM products');
    console.log('Products count:', r[0].cnt);
    const [c] = await conn.execute('SELECT COUNT(*) as cnt FROM categories');
    console.log('Categories count:', c[0].cnt);
    // Test a product query
    const [p] = await conn.execute('SELECT id, name, brand FROM products LIMIT 3');
    console.log('Sample products:', JSON.stringify(p, null, 2));
    await conn.end();
  } catch(e) {
    console.error('ERROR:', e.message);
    console.error('Code:', e.code);
  }
}
test();
