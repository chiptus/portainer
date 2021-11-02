import * as moment from 'moment-timezone';

export default class TimeWindowController {
  /* @ngInject */
  constructor($async) {
    this.$async = $async;
  }

  timeToUtc(startTimeSet, endTimeSet, timeZone) {
    const start = moment(startTimeSet).format('YYYY-MM-DD HH:mm');
    const end = moment(endTimeSet).format('YYYY-MM-DD HH:mm');

    const startTimeUtc = moment.tz(start, timeZone).utc().format('HH:mm');
    const endTimeUtc = moment.tz(end, timeZone).utc().format('HH:mm');

    return { startTimeUtc, endTimeUtc };
  }

  utcToTime(utcTime) {
    const startTime = moment.tz(utcTime.StartTime, 'HH:mm', 'GMT').tz(this.state.timeZoneSelected).format('HH:mm');
    const endTime = moment.tz(utcTime.EndTime, 'HH:mm', 'GMT').tz(this.state.timeZoneSelected).format('HH:mm');

    const [startHour, startMin] = startTime.split(':');
    const [endHour, endMin] = endTime.split(':');

    const startTimeUser = new Date();
    const endTimeUser = new Date();

    startTimeUser.setHours(startHour);
    startTimeUser.setMinutes(startMin);

    endTimeUser.setHours(endHour);
    endTimeUser.setMinutes(endMin);

    return {
      startTime: startTimeUser,
      endTime: endTimeUser,
    };
  }

  loadUtcTime() {
    const startTime = this.timeWindow.StartTime;
    const endTime = this.timeWindow.EndTime;

    const { startTimeObject, endTimeObject } = this.customToTimeObject(startTime, endTime);

    this.state.utcStartTime = startTimeObject;
    this.state.utcEndTime = endTimeObject;
  }

  loadTimeWindow() {
    // get time window from api ( UTC time object )
    const timeWindow = this.timeWindow;

    // Recover user set time with user-set timezone from UTC time object
    const { startTime, endTime } = this.utcToTime(timeWindow);

    this.state.setStartTime = startTime;
    this.state.setEndTime = endTime;
  }

  timeWindowUpdate() {
    const { startTimeUtc, endTimeUtc } = this.timeToUtc(this.state.setStartTime, this.state.setEndTime, this.state.timeZoneSelected);

    if (this.state.setStartTime.getTime() === this.state.setEndTime.getTime()) {
      this.state.timeError = true;
    } else {
      this.state.timeError = false;
    }

    this.timeWindow = {
      Enabled: this.timeWindow.Enabled,
      StartTime: startTimeUtc,
      EndTime: endTimeUtc,
    };

    this.timeZone = this.state.timeZoneSelected;

    const { startTimeObject, endTimeObject } = this.customToTimeObject(startTimeUtc, endTimeUtc);

    this.state.utcStartTime = startTimeObject;
    this.state.utcEndTime = endTimeObject;
  }

  customToTimeObject(startTime, endTime) {
    const [startHour, startMin] = startTime.split(':');
    const [endHour, endMin] = endTime.split(':');

    const startTimeObject = new Date();
    const endTimeObject = new Date();

    startTimeObject.setHours(startHour);
    startTimeObject.setMinutes(startMin);

    endTimeObject.setHours(endHour);
    endTimeObject.setMinutes(endMin);

    return { startTimeObject, endTimeObject };
  }

  defaultTimeWindow() {
    const defaultStartTime = '00:00';
    const defaultEndTime = '00:00';

    this.state.setStartTime.setHours(0, 0, 0, 0);
    this.state.setEndTime.setHours(0, 0, 0, 0);

    this.timeWindow = {
      Enabled: this.timeWindow.Enabled,
      StartTime: defaultStartTime,
      EndTime: defaultEndTime,
    };
  }

  toggleChangeWindow() {
    this.state.changeWindow = !this.state.changeWindow;
    if (!this.state.changeWindow) {
      const { startTime, endTime } = this.utcToTime(this.state.initTime);
      this.state.setStartTime = startTime;
      this.state.setEndTime = endTime;
      this.state.timeError = false;
    }
  }

  $onInit() {
    const currentUserTimezone = moment.tz.guess();

    const countries = moment.tz.countries();
    const zones = new Set();
    for (const country of countries) {
      moment.tz.zonesForCountry(country).reduce((set, zone) => set.add(zone), zones);
    }
    this.timeZones = [...zones].sort();
    this.timeZones.push('UTC');

    this.state = {
      hstep: 1,
      mstep: 5,
      ismeridian: true,
      utcStartTime: '',
      utcEndTime: '',
      setStartTime: new Date(),
      setEndTime: new Date(),
      initTime: this.timeWindow,
      dst: moment().isDST(),
      changeWindow: false,
      timeSet: false,
      options: {
        timezones: this.timeZones,
      },
      timeError: false,
      timeZoneSelected: currentUserTimezone,
    };

    // StartTime & EndTime is not Null
    if (this.timeWindow.StartTime && this.timeWindow.EndTime) {
      if (this.timeWindow.StartTime !== '00:00' || this.timeWindow.EndTime !== '00:00') {
        this.loadUtcTime();
        this.loadTimeWindow();
        this.state.timeSet = true;
      } else {
        this.defaultTimeWindow();
        this.state.timeSet = false;
      }
    }
    // StartTime & EndTime is Null
    else {
      this.defaultTimeWindow();
    }
  }
}
