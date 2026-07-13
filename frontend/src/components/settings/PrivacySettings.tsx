import { Link } from "@tanstack/react-router";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldLabel,
  FieldTitle,
} from "@/components/ui/field";
import {
  DEFAULT_PRIVACY_FIELDS,
  type PrivacyFields,
} from "@/lib/ai/privacy";
import { useJsonSetting } from "./useJsonSetting";
import { SettingBlock } from "./SettingBlock";

function PrivacyToggle({
  id,
  title,
  description,
  checked,
  disabled,
  onCheckedChange,
}: {
  id: string;
  title: string;
  description: string;
  checked: boolean;
  disabled?: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg border border-border px-4 py-3">
      <div className="space-y-1">
        <FieldLabel htmlFor={id} className="text-sm font-medium">
          {title}
        </FieldLabel>
        <FieldDescription>{description}</FieldDescription>
      </div>
      <Checkbox
        id={id}
        checked={checked}
        disabled={disabled}
        onCheckedChange={(value) => onCheckedChange(value === true)}
      />
    </div>
  );
}

export function PrivacySettings() {
  const privacy = useJsonSetting<PrivacyFields>(
    "privacy.fields",
    DEFAULT_PRIVACY_FIELDS,
  );

  const updateField = (field: keyof PrivacyFields, enabled: boolean) => {
    privacy.setValue({
      ...privacy.value,
      [field]: enabled,
    });
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <SettingBlock
        title="Privacy"
        description="Control what event data cloud models receive. Local models always get full context on your device."
      >
        <Field>
          <FieldTitle>Data shared with cloud models</FieldTitle>
          <FieldDescription>
            Only applies when a cloud endpoint is selected. Category names are
            always sent so the model can pick from your list.{" "}
            <Link to="/settings/ai" className="text-primary underline-offset-4 hover:underline">
              AI settings
            </Link>
          </FieldDescription>
        </Field>

        <div className="space-y-2">
          <PrivacyToggle
            id="privacy-title"
            title="Event titles"
            description="Required for useful suggestions."
            checked={privacy.value.title}
            onCheckedChange={(checked) => updateField("title", checked)}
          />
          <PrivacyToggle
            id="privacy-attendees"
            title="Attendee domains"
            description="Domains only — never email addresses."
            checked={privacy.value.attendees}
            onCheckedChange={(checked) => updateField("attendees", checked)}
          />
          <PrivacyToggle
            id="privacy-description"
            title="Event descriptions"
            description="Higher leak risk, low added signal."
            checked={privacy.value.description}
            onCheckedChange={(checked) => updateField("description", checked)}
          />
          <PrivacyToggle
            id="privacy-location"
            title="Event location"
            description="Optional context for categorization."
            checked={privacy.value.location}
            onCheckedChange={(checked) => updateField("location", checked)}
          />
        </div>
      </SettingBlock>
    </div>
  );
}
