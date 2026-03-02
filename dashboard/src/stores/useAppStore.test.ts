import { describe, it, expect, beforeEach } from "vitest";
import { useAppStore } from "./useAppStore";

describe("useAppStore", () => {
  beforeEach(() => {
    // Reset store between tests
    useAppStore.setState({
      profiles: [],
      profilesLoading: false,
      instances: [],
      instancesLoading: false,
      tabsChartData: [],
      currentTabs: {},
      agents: [],
      selectedAgentId: null,
      events: [],
      eventFilter: "all",
      serverInfo: null,
    });
  });

  describe("profiles", () => {
    it("sets profiles", () => {
      const profiles = [{ name: "test", id: "prof_123" }] as any;
      useAppStore.getState().setProfiles(profiles);
      expect(useAppStore.getState().profiles).toEqual(profiles);
    });

    it("sets profiles loading state", () => {
      useAppStore.getState().setProfilesLoading(true);
      expect(useAppStore.getState().profilesLoading).toBe(true);
    });
  });

  describe("instances", () => {
    it("sets instances", () => {
      const instances = [{ id: "inst_123", profileName: "test" }] as any;
      useAppStore.getState().setInstances(instances);
      expect(useAppStore.getState().instances).toEqual(instances);
    });
  });

  describe("chart data", () => {
    it("adds chart data point", () => {
      const point = { timestamp: Date.now(), inst_123: 5 };
      useAppStore.getState().addChartDataPoint(point);
      expect(useAppStore.getState().tabsChartData).toHaveLength(1);
      expect(useAppStore.getState().tabsChartData[0]).toEqual(point);
    });

    it("keeps only last 60 data points", () => {
      // Add 65 points
      for (let i = 0; i < 65; i++) {
        useAppStore.getState().addChartDataPoint({ timestamp: i, inst_123: i });
      }

      const data = useAppStore.getState().tabsChartData;
      expect(data).toHaveLength(60);
      // Should have points 5-64 (the last 60)
      expect(data[0].timestamp).toBe(5);
      expect(data[59].timestamp).toBe(64);
    });

    it("sets current tabs", () => {
      const tabs = {
        inst_123: [{ id: "tab_1", url: "https://example.com" }],
      } as any;
      useAppStore.getState().setCurrentTabs(tabs);
      expect(useAppStore.getState().currentTabs).toEqual(tabs);
    });
  });

  describe("events", () => {
    it("adds event to beginning of list", () => {
      const event1 = { type: "action", timestamp: "2024-01-01" } as any;
      const event2 = { type: "action", timestamp: "2024-01-02" } as any;

      useAppStore.getState().addEvent(event1);
      useAppStore.getState().addEvent(event2);

      const events = useAppStore.getState().events;
      expect(events[0]).toEqual(event2); // Most recent first
      expect(events[1]).toEqual(event1);
    });

    it("limits events to 100", () => {
      // Add 105 events
      for (let i = 0; i < 105; i++) {
        useAppStore.getState().addEvent({ type: "action", id: i } as any);
      }

      const events = useAppStore.getState().events;
      expect(events).toHaveLength(100);
      // Most recent should be id: 104
      expect(events[0].id).toBe(104);
    });

    it("clears events", () => {
      useAppStore.getState().addEvent({ type: "action" } as any);
      useAppStore.getState().clearEvents();
      expect(useAppStore.getState().events).toHaveLength(0);
    });

    it("sets event filter", () => {
      useAppStore.getState().setEventFilter("errors");
      expect(useAppStore.getState().eventFilter).toBe("errors");
    });
  });

  describe("agents", () => {
    it("sets agents", () => {
      const agents = [{ id: "agent_1", name: "Test Agent" }] as any;
      useAppStore.getState().setAgents(agents);
      expect(useAppStore.getState().agents).toEqual(agents);
    });

    it("sets selected agent id", () => {
      useAppStore.getState().setSelectedAgentId("agent_1");
      expect(useAppStore.getState().selectedAgentId).toBe("agent_1");
    });

    it("clears selected agent id", () => {
      useAppStore.getState().setSelectedAgentId("agent_1");
      useAppStore.getState().setSelectedAgentId(null);
      expect(useAppStore.getState().selectedAgentId).toBeNull();
    });
  });

  describe("settings", () => {
    it("has default settings", () => {
      const settings = useAppStore.getState().settings;
      expect(settings.stealth).toBe("light");
      expect(settings.screencast?.fps).toBe(1);
    });

    it("updates settings", () => {
      const newSettings = {
        screencast: { fps: 5, quality: 50, maxWidth: 1024 },
        stealth: "strict" as const,
        browser: { blockImages: true, blockMedia: true, noAnimations: true },
      };
      useAppStore.getState().setSettings(newSettings);
      expect(useAppStore.getState().settings).toEqual(newSettings);
    });
  });
});
