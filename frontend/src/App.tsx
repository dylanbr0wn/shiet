import "@/index.css";
import { AppShell } from "@/components/app-shell";
import { SchedulePage } from "@/components/schedule/SchedulePage";

function App() {
  return (
    <AppShell>
      {(titlebarPaddingClass) => (
        <SchedulePage titlebarPaddingClass={titlebarPaddingClass} />
      )}
    </AppShell>
  );
}

export default App;
