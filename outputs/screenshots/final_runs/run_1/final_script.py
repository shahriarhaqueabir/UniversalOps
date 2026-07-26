import asyncio
from playwright.async_api import async_playwright

async def main():
    async with async_playwright() as p:
        browser = await p.firefox.launch(headless=True)
        context = await browser.new_context(viewport={"width": 1280, "height": 1800})
        page = await context.new_page()

        log_file = "outputs/screenshots/final_runs/run_1/final_script_log.txt"
        with open(log_file, "w") as f:
            f.write("step 0 action: starting screenshot capture\n")

        def log_step(step, action):
            with open(log_file, "a") as f:
                f.write(f"step {step} action: {action}\n")

        # 1. Dashboard
        log_step(1, "visiting dashboard")
        await page.goto("http://localhost:5173/")
        await page.wait_for_timeout(5000) # Wait for initial load and possible onboarding modal

        # Check if onboarding modal is visible and close it if necessary
        onboarding_btn = page.locator("button:has-text('Get Started')")
        if await onboarding_btn.is_visible():
            log_step(1, "onboarding modal visible, clicking get started")
            await onboarding_btn.click()
            await page.wait_for_timeout(2000)

        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_1_dashboard.png")

        # 2. Reports
        log_step(2, "navigating to reports")
        await page.click('[data-automation-id="main-tab-reports"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_2_reports.png")

        # 3. SysOps
        log_step(3, "navigating to sysops")
        await page.click('[data-automation-id="main-tab-sysops"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_3_sysops.png")

        # 4. NetOps
        log_step(4, "navigating to netops")
        await page.click('[data-automation-id="main-tab-netops"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_4_netops.png")

        # 5. SecOps
        log_step(5, "navigating to secops")
        await page.click('[data-automation-id="main-tab-secops"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_5_secops.png")

        # 6. DevOps
        log_step(6, "navigating to devops")
        await page.click('[data-automation-id="main-tab-devops"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_6_devops.png")

        # 7. AIOps (Hawk)
        log_step(7, "navigating to aiops")
        await page.click('[data-automation-id="main-tab-aiops"]')
        await page.wait_for_timeout(2000)
        await page.screenshot(path="outputs/screenshots/final_runs/run_1/screenshots/final_execution_7_aiops.png")

        log_step(8, "finished all screenshots")
        await browser.close()

if __name__ == "__main__":
    asyncio.run(main())
