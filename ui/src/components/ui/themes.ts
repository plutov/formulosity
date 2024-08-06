import { CustomFlowbiteTheme } from 'flowbite-react'

export const DatepicketTheme: CustomFlowbiteTheme['datepicker'] = {
  popup: {
    footer: {
      button: {
        base: 'w-full rounded-lg px-5 py-2 text-center text-sm font-medium focus:ring-4',
        today: 'datepicker-selected',
      },
    },
  },
  views: {
    days: {
      items: {
        item: {
          selected: 'datepicker-selected',
        },
      },
    },
    months: {
      items: {
        item: {
          selected: 'datepicker-selected',
        },
      },
    },
    years: {
      items: {
        item: {
          selected: 'datepicker-selected',
        },
      },
    },
    decades: {
      items: {
        item: {
          selected: 'datepicker-selected',
        },
      },
    },
  },
}
