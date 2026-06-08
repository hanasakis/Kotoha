const fs = require('fs');

async function main() {
  // Get embedding
  const embRes = await fetch('http://localhost:11434/api/embeddings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ model: 'quentinz/bge-large-zh-v1.5', prompt: '坚果 杏仁 零食' })
  });
  const embData = await embRes.json();
  const vec = embData.embedding;
  console.log('Embedding dim:', vec.length);

  // Search Milvus
  const searchBody = {
    collectionName: 'kotoha_products',
    dbName: 'default',
    data: [vec],
    annsField: 'vector',
    limit: 5,
    outputFields: ['id']
  };
  const res = await fetch('http://localhost:19530/v2/vectordb/entities/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer cm9vdDpNaWx2dXM='
    },
    body: JSON.stringify(searchBody)
  });
  const data = await res.json();
  console.log('Status:', res.status);
  console.log('Response:', JSON.stringify(data, null, 2));
}

main().catch(e => console.error(e));
