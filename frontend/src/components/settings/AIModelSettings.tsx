import {
  Check,
  LoaderCircle,
  RefreshCw,
  ShieldCheck,
  ShieldAlert,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useClassifyAIEndpoint,
  useDiscoverLocalAIEndpoints,
  useListAIModels,
  useSaveAIConfig,
  useSetting,
  useValidateAIConfig,
} from "@/lib/api";
import { SettingBlock } from "./SettingBlock";

function parseJsonSetting<T>(raw: string | null | undefined, fallback: T): T {
  if (!raw) {
    return fallback;
  }

  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export function AIModelSettings() {
  const baseURLSetting = useSetting("ai.base_url");
  const modelSetting = useSetting("ai.model");
  const savedBaseURL = useMemo(
    () => parseJsonSetting(baseURLSetting.data, ""),
    [baseURLSetting.data],
  );
  const savedModel = useMemo(
    () => parseJsonSetting(modelSetting.data, ""),
    [modelSetting.data],
  );

  const [baseURL, setBaseURL] = useState(savedBaseURL);
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState(savedModel);
  const [models, setModels] = useState<string[]>([]);
  const [validationMessage, setValidationMessage] = useState<string | null>(
    null,
  );

  const discovery = useDiscoverLocalAIEndpoints();
  const classify = useClassifyAIEndpoint(baseURL);
  const listModels = useListAIModels();
  const validate = useValidateAIConfig();
  const saveConfig = useSaveAIConfig();

  useEffect(() => {
    setBaseURL(savedBaseURL);
  }, [savedBaseURL]);

  useEffect(() => {
    setModel(savedModel);
  }, [savedModel]);

  const classification = useMemo(() => {
    if (!baseURL.trim()) {
      return null;
    }
    return classify.data ?? null;
  }, [baseURL, classify.data]);

  const refreshModels = async () => {
    if (!baseURL.trim()) {
      return;
    }

    const nextModels = await listModels.mutateAsync({
      baseURL,
      apiKey,
    });
    setModels(nextModels);
    if (nextModels.length > 0 && !nextModels.includes(model)) {
      setModel(nextModels[0]);
    }
  };

  const handleSelectEndpoint = async (nextBaseURL: string, nextModel?: string) => {
    setBaseURL(nextBaseURL);
    if (nextModel) {
      setModel(nextModel);
    }
    const nextModels = await listModels.mutateAsync({
      baseURL: nextBaseURL,
      apiKey,
    });
    setModels(nextModels);
    if (!nextModel && nextModels.length > 0) {
      setModel(nextModels[0]);
    }
  };

  const handleValidate = async () => {
    const result = await validate.mutateAsync({
      baseURL,
      apiKey,
      model,
    });
    setValidationMessage(result.message);
    if (result.ok) {
      await saveConfig.mutateAsync({ baseURL, model });
    }
  };

  const isBusy =
    discovery.isLoading ||
    listModels.isPending ||
    validate.isPending ||
    saveConfig.isPending;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <SettingBlock
        title="AI Model"
        description="Bring your own model. Point Clockr at a local runtime or a custom OpenAI-compatible endpoint for categorization suggestions."
      >
        <div className="space-y-2">
          <div className="flex items-center justify-between gap-3">
            <Label className="text-xs">Detected endpoints</Label>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              disabled={discovery.isFetching}
              onClick={() => void discovery.refetch()}
            >
              <RefreshCw
                className={`size-4 ${discovery.isFetching ? "animate-spin" : ""}`}
              />
              Scan
            </Button>
          </div>

          <div className="space-y-2">
            {(discovery.data ?? []).map((endpoint) => (
              <button
                key={endpoint.baseUrl}
                type="button"
                className={`flex w-full items-center justify-between rounded-lg border px-3 py-2 text-left transition-colors ${
                  baseURL === endpoint.baseUrl
                    ? "border-primary bg-primary/5"
                    : "border-border hover:bg-muted/50"
                }`}
                onClick={() =>
                  void handleSelectEndpoint(
                    endpoint.baseUrl,
                    endpoint.models?.[0],
                  )
                }
              >
                <div>
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <span>{endpoint.name}</span>
                    <span
                      className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${
                        endpoint.running
                          ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      {endpoint.running ? "Running" : "Not running"}
                    </span>
                  </div>
                  <p className="mt-0.5 font-mono text-xs text-muted-foreground">
                    {endpoint.baseUrl}
                  </p>
                </div>
              </button>
            ))}
          </div>
        </div>
      </SettingBlock>

      <SettingBlock
        title="Connection"
        description="Configure the OpenAI-compatible base URL and model Clockr should use."
      >
        <div className="grid gap-3">
          <div className="grid gap-1.5">
            <Label htmlFor="ai-base-url" className="text-xs">
              Base URL
            </Label>
            <Input
              id="ai-base-url"
              className="font-mono"
              value={baseURL}
              onChange={(event) => setBaseURL(event.target.value)}
              placeholder="http://127.0.0.1:11434/v1"
            />
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="ai-api-key" className="text-xs">
              API key <span className="text-muted-foreground">(optional)</span>
            </Label>
            <Input
              id="ai-api-key"
              type="password"
              className="font-mono"
              value={apiKey}
              onChange={(event) => setApiKey(event.target.value)}
              placeholder="Not required for local models"
            />
          </div>

          <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
            <div className="grid gap-1.5">
              <Label htmlFor="ai-model" className="text-xs">
                Model
              </Label>
              <Select value={model} onValueChange={setModel}>
                <SelectTrigger id="ai-model" className="w-full bg-background">
                  <SelectValue placeholder="Select or type a model" />
                </SelectTrigger>
                <SelectContent>
                  {models.map((item) => (
                    <SelectItem key={item} value={item}>
                      {item}
                    </SelectItem>
                  ))}
                  {model && !models.includes(model) ? (
                    <SelectItem value={model}>{model}</SelectItem>
                  ) : null}
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                disabled={!baseURL.trim() || isBusy}
                onClick={() => void refreshModels()}
              >
                {listModels.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : (
                  "Load models"
                )}
              </Button>
              <Button
                type="button"
                disabled={!baseURL.trim() || !model.trim() || isBusy}
                onClick={() => void handleValidate()}
              >
                {validate.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : (
                  <Check className="size-4" />
                )}
                Validate
              </Button>
            </div>
          </div>

          <Input
            aria-label="Custom model name"
            className="font-mono"
            value={model}
            onChange={(event) => setModel(event.target.value)}
            placeholder="Or type a model name"
          />

          {classification ? (
            <div
              className={`flex gap-3 rounded-lg border px-3 py-3 text-sm ${
                classification.local
                  ? "border-emerald-500/30 bg-emerald-500/5"
                  : "border-amber-500/30 bg-amber-500/5"
              }`}
            >
              {classification.local ? (
                <ShieldCheck className="mt-0.5 size-4 shrink-0 text-emerald-600" />
              ) : (
                <ShieldAlert className="mt-0.5 size-4 shrink-0 text-amber-600" />
              )}
              <div>
                <p className="font-medium">
                  {classification.local
                    ? "Private — on-device"
                    : "Cloud — data may leave your device"}
                </p>
                <p className="mt-1 text-muted-foreground">
                  {classification.verdict}
                </p>
              </div>
            </div>
          ) : null}

          {validationMessage ? (
            <p className="text-sm text-muted-foreground">{validationMessage}</p>
          ) : null}
        </div>
      </SettingBlock>
    </div>
  );
}
