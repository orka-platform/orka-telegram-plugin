import { generateText } from 'ai';
import { createOpenAI } from '@ai-sdk/openai';
import { createAnthropic } from '@ai-sdk/anthropic';
import { createGoogleGenerativeAI } from '@ai-sdk/google';

function readStdin() {
  return new Promise((resolve, reject) => {
    try {
      const chunks = [];
      process.stdin.on('data', (c) => chunks.push(Buffer.isBuffer(c) ? c : Buffer.from(c)));
      process.stdin.on('end', () => {
        const raw = Buffer.concat(chunks).toString('utf8');
        try { resolve(JSON.parse(raw)); } catch (e) { reject(e); }
      });
      process.stdin.on('error', reject);
    } catch (err) {
      reject(err);
    }
  });
}

function normalizeMessages(input) {
  if (!Array.isArray(input)) return [];
  return input.map((m) => {
    if (m && typeof m === 'object') {
      const role = typeof m.role === 'string' ? m.role : 'user';
      const content = typeof m.content === 'string' ? m.content : String(m.content ?? '');
      return { role, content };
    }
    return { role: 'user', content: String(m) };
  });
}

function getProvider({ provider, apiKey, baseURL }) {
  switch ((provider || '').toLowerCase()) {
    case 'openai':
      return createOpenAI({ apiKey, baseURL: baseURL || undefined });
    case 'anthropic':
      return createAnthropic({ apiKey, baseURL: baseURL || undefined });
    case 'google':
    case 'gemini':
      return createGoogleGenerativeAI({ apiKey, baseURL: baseURL || undefined });
    default:
      throw new Error(`unsupported provider: ${provider}`);
  }
}

(async () => {
  try {
    const input = await readStdin();
    const {
      provider,
      model: modelId,
      apiKey,
      baseURL,
      messages: rawMessages,
      system,
      temperature,
      maxTokens,
    } = input || {};

    if (!provider || !modelId || !apiKey || !rawMessages) {
      console.error('missing required inputs');
      process.stdout.write(JSON.stringify({ success: false, error: 'missing required inputs' }));
      return;
    }

    const providerFactory = getProvider({ provider, apiKey, baseURL });
    const model = providerFactory(modelId);

    const messages = normalizeMessages(rawMessages);
    if (system && typeof system === 'string') {
      messages.unshift({ role: 'system', content: system });
    }

    const result = await generateText({
      model,
      messages,
      temperature: typeof temperature === 'number' ? temperature : undefined,
      maxTokens: typeof maxTokens === 'number' ? maxTokens : undefined,
    });

    const out = {
      success: true,
      text: result?.text ?? '',
      model: result?.modelId || modelId,
      finishReason: result?.finishReason || 'stop',
      usage: result?.usage || undefined,
      extra: undefined,
    };
    process.stdout.write(JSON.stringify(out));
  } catch (err) {
    process.stdout.write(JSON.stringify({ success: false, error: String(err?.message || err) }));
  }
})();